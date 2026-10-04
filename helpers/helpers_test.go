package helpers

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setVersion swaps the link-time variable for the duration of a test.
func setVersion(t *testing.T, v string) {
	t.Helper()

	previous := version
	version = v

	t.Cleanup(func() { version = previous })
}

// markRoot creates a project marker at dir and returns dir.
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
		t.Fatalf("esperava %q, obteve %q", want, got)
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
		t.Fatalf("esperava %q, obteve %q", want, got)
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
		t.Fatalf("esperava %q, obteve %q", root, got)
	}
}

func TestFindProjectRootReportsWhereItStarted(t *testing.T) {
	// A temp dir has no go.mod above it, so the walk runs to the filesystem
	// root and fails. The message has to name the directory it started from.
	start := t.TempDir()

	_, err := findProjectRoot(start)

	if err == nil {
		t.Fatal("esperava erro ao não achar a raiz do projeto")
	}

	if !strings.Contains(err.Error(), start) {
		t.Fatalf("a mensagem deveria citar %q, obteve %q", start, err)
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
		t.Fatalf("um diretório chamado %s não vale como raiz: esperava %q, obteve %q", projectMarker, root, got)
	}
}

func TestFindProjectRootPropagatesUnexpectedStatError(t *testing.T) {
	// Pointing at a regular file makes os.Stat return ENOTDIR for
	// <file>/go.mod, which is not ErrNotExist and must not be mistaken for
	// "keep walking up".
	blocker := filepath.Join(t.TempDir(), "blocker")

	if err := os.WriteFile(blocker, nil, 0644); err != nil {
		t.Fatalf("writing blocker: %v", err)
	}

	_, err := findProjectRoot(blocker)

	if err == nil {
		t.Fatal("esperava propagação do erro de stat")
	}

	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("o erro não deveria parecer ausência do marcador: %v", err)
	}

	if !strings.Contains(err.Error(), "checking for") {
		t.Fatalf("erro inesperado, sem contexto: %v", err)
	}
}
