package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

func StateDir() (string, error) {
	if dir := os.Getenv("RUNNER_STATE_DIR"); dir != "" {
		return dir, nil
	}

	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve state dir: %w", err)
	}

	return filepath.Join(base, "runner"), nil
}

func VMDir(id string) (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vms", id), nil
}
