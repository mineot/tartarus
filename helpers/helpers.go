package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const projectMarker = "go.mod"

func IsProduction() bool {
	exePath, err := os.Executable()

	if err != nil {
		return true
	}

	return !strings.Contains(exePath, "go-build")
}

func GetProductionStorePath(dbName string) (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	path := homeDir + "/.tartarus/" + dbName

	return path, nil
}

func GetDevelopmentStorePath(dbName string) (string, error) {
	dir, err := os.Getwd()

	if err != nil {
		return "", err
	}

	root, err := findProjectRoot(dir)

	if err != nil {
		return "", err
	}

	return filepath.Join(root, dbName), nil
}

func findProjectRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, projectMarker)); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			return "", fmt.Errorf("project root not found: no %s in any parent directory of %s", projectMarker, dir)
		}

		dir = parent
	}
}
