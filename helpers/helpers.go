package helpers

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// version is injected at link time by the Makefile, through
// -ldflags "-X tartarus/helpers.version=...". A plain "go run ." and a bare
// "go build ." both leave it at devVersion, which means development.
//
// It has to stay a string with a constant initializer: the linker refuses to
// rewrite anything else.
var version = "dev"

const (
	// devVersion is the value injected by the Makefile.
	devVersion = "dev"
	// productionDir is the directory under which the production database is
	productionDir = ".tartarus"
	// productionDB is the name of the production database
	productionDB = "tartarus.db"
	// developmentDB is the name of the development database
	developmentDB = "dev.db"
	// projectMarker is the marker for the project root
	projectMarker = "go.mod"
)

// GetStorePath returns the database path for the current build: the production
// database under ~/.tartarus, the development one at the project root.
func GetStorePath() (string, error) {
	if isProduction() {
		return productionPath()
	}

	return developmentPath()
}

// ErrProduction is returned by SetDevStorePath in a build that has a version
// injected by the Makefile.
var ErrProduction = errors.New("helpers: the store path is fixed in a production build")

// devStorePath overrides the development database when it is not empty.
//
// It exists so tests in other packages can reach a database at all: store's
// path-taking constructor is unexported, so store.New, and therefore this
// function, is the only door.
var devStorePath string

// SetDevStorePath points the development database at path. An empty path clears
// the override and goes back to resolving the project root.
//
// It only takes effect in a development build. Once a version is injected, it
// returns ErrProduction and the path stays the production one, so a released
// binary cannot be made to open some other file.
//
// The override is a plain package variable with no locking. Call it before
// starting concurrent work, which in practice means from a test helper.
func SetDevStorePath(path string) error {
	if isProduction() {
		return ErrProduction
	}

	devStorePath = path

	return nil
}

// isProduction returns true if the current build is a production build.
func isProduction() bool {
	return version != devVersion
}

// productionPath returns the path to the production database.
func productionPath() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	return filepath.Join(home, productionDir, productionDB), nil
}

// developmentPath returns the path to the development database.
func developmentPath() (string, error) {
	// The override short-circuits the search for the project root: it exists to
	// point a test at a temp directory, which is not inside a project at all.
	if devStorePath != "" {
		return devStorePath, nil
	}

	wd, err := os.Getwd()

	if err != nil {
		return "", fmt.Errorf("resolving working directory: %w", err)
	}

	root, err := findProjectRoot(wd)

	if err != nil {
		return "", err
	}

	return filepath.Join(root, developmentDB), nil
}

// findProjectRoot walks up from dir looking for the project marker. It reports
// the directory it started from when it fails, not the filesystem root it ended
// up at.
func findProjectRoot(dir string) (string, error) {
	start := dir

	for {
		info, err := os.Stat(filepath.Join(dir, projectMarker))

		switch {
		case err == nil && !info.IsDir():
			return dir, nil
		case err != nil && !errors.Is(err, fs.ErrNotExist):
			return "", fmt.Errorf("checking for %s in %s: %w", projectMarker, dir, err)
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			return "", fmt.Errorf(
				"project root not found: no %s in any parent directory of %s",
				projectMarker,
				start,
			)
		}

		dir = parent
	}
}
