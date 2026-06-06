package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	KernelImage = "vmlinux"
	InitrdImage = "initrd"
	RootfsImage = "rootfs.ext4"
)

var required = []string{
	KernelImage,
	InitrdImage,
	RootfsImage,
}

func ResolveDir() (string, error) {
	if dir := os.Getenv("RUNNER_ASSETS_DIR"); dir != "" {
		return dir, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve assets dir: %w", err)
	}

	return filepath.Dir(exe), nil
}

func Validate(dir string) error {
	for _, name := range required {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("missing asset %q in %s: %w", name, dir, err)
		}
	}
	return nil
}

func Symlink(srcDir, vmDir, name string) error {
	target := filepath.Join(srcDir, name)
	link := filepath.Join(vmDir, name)
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("link %s: %w", name, err)
	}
	return nil
}
