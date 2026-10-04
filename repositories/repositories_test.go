package repositories

import (
	"context"
	"path/filepath"
	"tartarus/helpers"
	"tartarus/store"
	"testing"
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
