package repositories

import (
	"context"
	"path/filepath"
	"tartarus/helpers"
	"tartarus/store"
	"testing"
)

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
