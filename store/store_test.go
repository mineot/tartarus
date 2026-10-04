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
		t.Fatalf("esperava 2 itens confirmados, obteve %d", got)
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
		t.Fatalf("WithTx deveria devolver %v, obteve %v", failure, err)
	}

	if got := count(t, s); got != 0 {
		t.Fatalf("rollback deveria ter zerado a tabela, obteve %d", got)
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
			return errors.New("Exec do Store dentro de WithTx deveria devolver ErrUseTx")
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
		t.Fatalf("WithTx aninhado deveria devolver ErrTxActive, obteve %v", err)
	}
}

func TestOperationsAfterClose(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close deveria ser idempotente, obteve %v", err)
	}

	if _, err := s.Exec(`INSERT INTO items (name) VALUES (?)`, "a"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Exec após Close deveria devolver ErrClosed, obteve %v", err)
	}

	if err := s.WithTx(func(tx *Tx) error { return nil }); !errors.Is(err, ErrClosed) {
		t.Fatalf("WithTx após Close deveria devolver ErrClosed, obteve %v", err)
	}
}

func TestWithTxWithoutFunc(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	if err := s.WithTx(nil); !errors.Is(err, ErrNoFunc) {
		t.Fatalf("WithTx(nil) deveria devolver ErrNoFunc, obteve %v", err)
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
		t.Fatalf("esperava %q, obteve %q", want, got)
	}
}

func TestPanicInsideWithTxReleasesConnection(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	func() {
		defer func() {
			if recover() == nil {
				t.Error("o pânico de dentro de WithTx deveria propagar")
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
			t.Fatalf("Store travou depois do pânico: %v", err)
		}

		if got := count(t, s); got != 1 {
			t.Fatalf("apenas o insert fora da transação deveria ter sobrado, obteve %d", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Store travou: a conexão não foi devolvida após o pânico")
	}
}
