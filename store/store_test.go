package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

const createItems = `CREATE TABLE IF NOT EXISTS items (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`

func newTestStore(t *testing.T) *Store {
	t.Helper()

	s, err := newAt(context.Background(), filepath.Join(t.TempDir(), "test.db"))

	if err != nil {
		t.Fatalf("newAt: %v", err)
	}

	if _, err = s.Exec(createItems); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return s
}

func count(t *testing.T, s *Store) int {
	t.Helper()

	var n int

	if err := s.db.QueryRowContext(s.ctx, `SELECT COUNT(*) FROM items`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}

	return n
}

func TestWithTxCommit(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	err := s.WithTx(func(tx *Tx) error {
		if _, err := tx.Exec(`INSERT INTO items (name) VALUES (?)`, "a"); err != nil {
			return err
		}

		_, err := tx.Exec(`INSERT INTO items (name) VALUES (?)`, "b")

		return err
	})

	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	if got := count(t, s); got != 2 {
		t.Fatalf("want 2 committed items, got %d", got)
	}
}

func TestWithTxRollback(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	failure := context.Canceled

	err := s.WithTx(func(tx *Tx) error {
		if _, err := tx.Exec(`INSERT INTO items (name) VALUES (?)`, "a"); err != nil {
			return err
		}

		return failure
	})

	if !errors.Is(err, failure) {
		t.Fatalf("WithTx should return %v, got %v", failure, err)
	}

	if got := count(t, s); got != 0 {
		t.Fatalf("rollback should have emptied the table, got %d", got)
	}
}

func TestExecDuringTxIsRejected(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	err := s.WithTx(func(tx *Tx) error {
		if _, err := tx.Exec(`INSERT INTO items (name) VALUES (?)`, "a"); err != nil {
			return err
		}

		if _, err := s.Exec(`INSERT INTO items (name) VALUES (?)`, "b"); !errors.Is(err, ErrUseTx) {
			return errors.New("Store.Exec inside WithTx should return ErrUseTx")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}
}

func TestNestedWithTxIsRejected(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	err := s.WithTx(func(tx *Tx) error {
		return s.WithTx(func(tx *Tx) error {
			return nil
		})
	})

	if !errors.Is(err, ErrTxActive) {
		t.Fatalf("nested WithTx should return ErrTxActive, got %v", err)
	}
}

func TestOperationsAfterClose(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close should be idempotent, got %v", err)
	}

	if _, err := s.Exec(`INSERT INTO items (name) VALUES (?)`, "a"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Exec after Close should return ErrClosed, got %v", err)
	}

	if err := s.WithTx(func(tx *Tx) error { return nil }); !errors.Is(err, ErrClosed) {
		t.Fatalf("WithTx after Close should return ErrClosed, got %v", err)
	}
}

func TestWithTxWithoutFunc(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	if err := s.WithTx(nil); !errors.Is(err, ErrNoFunc) {
		t.Fatalf("WithTx(nil) should return ErrNoFunc, got %v", err)
	}
}

func TestPath(t *testing.T) {
	want := filepath.Join(t.TempDir(), "test.db")

	s, err := newAt(context.Background(), want)

	if err != nil {
		t.Fatalf("newAt: %v", err)
	}

	defer s.Close()

	if got := s.Path(); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestPanicInsideWithTxReleasesConnection(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	func() {
		defer func() {
			if recover() == nil {
				t.Error("the panic from inside WithTx should propagate")
			}
		}()

		s.WithTx(func(tx *Tx) error {
			tx.Exec(`INSERT INTO items (name) VALUES (?)`, "a")

			panic("boom")
		})
	}()

	done := make(chan error, 1)

	go func() {
		_, err := s.Exec(`INSERT INTO items (name) VALUES (?)`, "b")
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Store locked up after the panic: %v", err)
		}

		if got := count(t, s); got != 1 {
			t.Fatalf("only the insert outside the transaction should have survived, got %d", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Store locked up: the connection was not returned after the panic")
	}
}
