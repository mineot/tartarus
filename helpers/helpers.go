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
	devVersion = "dev"

	productionDir = ".tartarus"
	productionDB  = "tartarus.db"
	developmentDB = "dev.db"

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

func isProduction() bool {
	return version != devVersion
}

func productionPath() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	return filepath.Join(home, productionDir, productionDB), nil
}

func developmentPath() (string, error) {
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
