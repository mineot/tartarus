package helpers

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setVersion(t *testing.T, v string) {
	t.Helper()

	previous := version
	version = v

	t.Cleanup(func() { version = previous })
}

func markRoot(t *testing.T, dir string) string {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, projectMarker), nil, 0644); err != nil {
		t.Fatalf("writing project marker: %v", err)
	}

	return dir
}

func TestIsProduction(t *testing.T) {
	t.Run("development by default", func(t *testing.T) {
		setVersion(t, devVersion)

		if isProduction() {
			t.Error("an uninjected build should be development")
		}
	})

	t.Run("production once a version is injected", func(t *testing.T) {
		setVersion(t, "v1.0.4")

		if !isProduction() {
			t.Error("an injected version should be production")
		}
	})
}

func TestGetStorePathProduction(t *testing.T) {
	setVersion(t, "v1.0.4")
	t.Setenv("HOME", t.TempDir())

	got, err := GetStorePath()

	if err != nil {
		t.Fatalf("GetStorePath: %v", err)
	}

	want := filepath.Join(os.Getenv("HOME"), productionDir, productionDB)

	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestGetStorePathDevelopment(t *testing.T) {
	setVersion(t, devVersion)

	root := markRoot(t, t.TempDir())
	t.Chdir(root)

	got, err := GetStorePath()

	if err != nil {
		t.Fatalf("GetStorePath: %v", err)
	}

	want := filepath.Join(root, developmentDB)

	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestFindProjectRootFromSubdirectory(t *testing.T) {
	root := markRoot(t, t.TempDir())
	nested := filepath.Join(root, "repositories")

	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := findProjectRoot(nested)

	if err != nil {
		t.Fatalf("findProjectRoot: %v", err)
	}

	if got != root {
		t.Fatalf("want %q, got %q", root, got)
	}
}

func TestFindProjectRootReportsWhereItStarted(t *testing.T) {
	start := t.TempDir()

	_, err := findProjectRoot(start)

	if err == nil {
		t.Fatal("want an error when the project root is not found")
	}

	if !strings.Contains(err.Error(), start) {
		t.Fatalf("the message should name %q, got %q", start, err)
	}
}

func TestFindProjectRootRejectsDirectoryMarker(t *testing.T) {
	root := markRoot(t, t.TempDir())
	decoy := filepath.Join(root, "nested")

	if err := os.MkdirAll(filepath.Join(decoy, projectMarker), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := findProjectRoot(decoy)

	if err != nil {
		t.Fatalf("findProjectRoot: %v", err)
	}

	if got != root {
		t.Fatalf("a directory named %s does not count as the root: want %q, got %q", projectMarker, root, got)
	}
}

func TestFindProjectRootPropagatesUnexpectedStatError(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")

	if err := os.WriteFile(blocker, nil, 0644); err != nil {
		t.Fatalf("writing blocker: %v", err)
	}

	_, err := findProjectRoot(blocker)

	if err == nil {
		t.Fatal("want the stat error to propagate")
	}

	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the error should not look like a missing marker: %v", err)
	}

	if !strings.Contains(err.Error(), "checking for") {
		t.Fatalf("unexpected error, no context: %v", err)
	}
}

func TestSetDevStorePath(t *testing.T) {
	setVersion(t, devVersion)

	want := filepath.Join(t.TempDir(), "override.db")

	if err := SetDevStorePath(want); err != nil {
		t.Fatalf("SetDevStorePath: %v", err)
	}

	t.Cleanup(func() { devStorePath = "" })

	got, err := GetStorePath()

	if err != nil {
		t.Fatalf("GetStorePath: %v", err)
	}

	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestSetDevStorePathEmptyClearsTheOverride(t *testing.T) {
	setVersion(t, devVersion)

	root := markRoot(t, t.TempDir())
	t.Chdir(root)

	if err := SetDevStorePath(filepath.Join(t.TempDir(), "override.db")); err != nil {
		t.Fatalf("SetDevStorePath: %v", err)
	}

	t.Cleanup(func() { devStorePath = "" })

	if err := SetDevStorePath(""); err != nil {
		t.Fatalf("SetDevStorePath with an empty string: %v", err)
	}

	got, err := GetStorePath()

	if err != nil {
		t.Fatalf("GetStorePath: %v", err)
	}

	if want := filepath.Join(root, developmentDB); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestSetDevStorePathIsRejectedInProduction(t *testing.T) {
	setVersion(t, "v1.0.4")
	t.Setenv("HOME", t.TempDir())

	err := SetDevStorePath(filepath.Join(t.TempDir(), "override.db"))

	if !errors.Is(err, ErrProduction) {
		t.Fatalf("want ErrProduction, got %v", err)
	}

	if devStorePath != "" {
		t.Fatalf("the override should not have been written, got %q", devStorePath)
	}

	want := filepath.Join(os.Getenv("HOME"), productionDir, productionDB)

	if got, err := GetStorePath(); err != nil {
		t.Fatalf("GetStorePath: %v", err)
	} else if got != want {
		t.Fatalf("the production path changed: want %q, got %q", want, got)
	}
}
