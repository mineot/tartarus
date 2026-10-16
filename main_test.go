package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"tartarus/helpers"
)

func captureStdout(t *testing.T, fn func() error) error {
	t.Helper()

	original := os.Stdout

	r, w, err := os.Pipe()

	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdout = w

	fnErr := fn()

	os.Stdout = original

	if err := w.Close(); err != nil {
		t.Fatalf("closing the pipe writer: %v", err)
	}

	if _, err := io.ReadAll(r); err != nil {
		t.Fatalf("reading the pipe: %v", err)
	}

	if err := r.Close(); err != nil {
		t.Fatalf("closing the pipe reader: %v", err)
	}

	return fnErr
}

func TestRunExecutesTheRootCommand(t *testing.T) {
	if err := helpers.SetDevStorePath(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("SetDevStorePath: %v", err)
	}

	originalArgs := os.Args
	os.Args = []string{"tartarus"}

	t.Cleanup(func() { os.Args = originalArgs })

	err := captureStdout(t, run)

	if err != nil {
		t.Fatalf("run: %v", err)
	}
}
