package repositories

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"tartarus/helpers"
	"tartarus/store"
)

// newTestStore points the development database at a temp directory and returns a
// migrated Store.
//
// It goes through helpers.SetDevStorePath instead of a path argument because the
// store's path-taking constructor is unexported, leaving store.New as the only
// door into a database from out here. Each call gets its own file, so tests do
// not share state and do not need a reset in between.
//
// helpers.SetDevStorePath writes a package level variable with no locking, so
// this must not be called from a test running in parallel.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	if err := helpers.SetDevStorePath(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("SetDevStorePath: %v", err)
	}

	s, err := store.New(context.Background())

	if err != nil {
		t.Fatalf("store.New: %v", err)
	}

	t.Cleanup(func() { s.Close() })

	if err = s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return s
}

// insertManual writes a manual and returns the timestamp it was stamped with,
// which the caller needs in order to compare what comes back out.
func insertManual(t *testing.T, s *store.Store, name string, body string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO manuals (name, body, created_at, updated_at) VALUES (?, ?, ?, ?)`, name, body, now, now)

		if err != nil {
			return err
		}

		id, err = result.LastInsertId()

		return err
	})

	if err != nil {
		t.Fatalf("inserting manual %q: %v", name, err)
	}

	return uint64(id), now
}

func TestGetManualsOnEmptyDatabase(t *testing.T) {
	s := newTestStore(t)

	manuals, err := GetManuals(s)

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 0 {
		t.Fatalf("esperava zero manuais, obteve %d", len(manuals))
	}
}

func TestGetManualsReturnsEveryManualInCreationOrder(t *testing.T) {
	s := newTestStore(t)

	type inserted struct {
		id   uint64
		name string
		body string
		at   time.Time
	}

	want := make([]inserted, 0, 3)

	// Inserted in a deliberate order that is not alphabetical, so a query that
	// sorted by name would fail the comparison below.
	for _, m := range []struct{ name, body string }{
		{"zebra", "last alphabetically"},
		{"alpha", "first alphabetically"},
		{"mango", "in between"},
	} {
		id, at := insertManual(t, s, m.name, m.body)
		want = append(want, inserted{id: id, name: m.name, body: m.body, at: at})
	}

	manuals, err := GetManuals(s)

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != len(want) {
		t.Fatalf("esperava %d manuais, obteve %d", len(want), len(manuals))
	}

	for i, w := range want {
		got := manuals[i]

		if got.ID != w.id {
			t.Fatalf("posição %d: esperava o id %d, obteve %d", i, w.id, got.ID)
		}

		if got.Name != w.name {
			t.Fatalf("posição %d: esperava o nome %q, obteve %q", i, w.name, got.Name)
		}

		if got.Body != w.body {
			t.Fatalf("posição %d: esperava o corpo %q, obteve %q", i, w.body, got.Body)
		}

		// Compared at second granularity: the round trip through SQLite goes
		// through a text representation, so the exact instant is not preserved
		// in a way worth asserting on.
		if got.CreatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava created_at %d, obteve %d", i, w.at.Unix(), got.CreatedAt.Unix())
		}

		if got.UpdatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava updated_at %d, obteve %d", i, w.at.Unix(), got.UpdatedAt.Unix())
		}
	}
}

// TestGetManualsIsRejectedInsideATransaction pins down the restriction the
// GetManuals doc comment promises: reads go through Store.Query, which refuses to
// run while the connection belongs to a transaction.
func TestGetManualsIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		_, err := GetManuals(s)

		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
	}
}

// TestGetManualsPropagatesAClosedStore covers the error path, which has to stay
// matchable with errors.Is so callers can tell it from a query failure.
func TestGetManualsPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := GetManuals(s)

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}
