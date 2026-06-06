package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	KernelImage           = "vmlinux"
	InitrdImage           = "initrd"
	RootfsImage           = "rootfs.ext4"
	RootfsImageCompressed = "rootfs.ext4.zst"
	BootArgsFile          = "boot.args"
)

var required = []string{
	KernelImage,
	InitrdImage,
	BootArgsFile,
}

func BootArgs(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, BootArgsFile))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", BootArgsFile, err)
	}

	args := strings.TrimSpace(string(data))
	if args == "" {
		return "", fmt.Errorf("%s is empty", BootArgsFile)
	}

	return args, nil
}

const (
	InstalledBinaryDir = "/usr/local/bin"
	InstalledAssetsDir = "/usr/local/share/runner"
)

func ResolveDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve assets dir: %w", err)
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolve assets dir: %w", err)
	}

	if filepath.Dir(exe) == InstalledBinaryDir {
		return InstalledAssetsDir, nil
	}

	return "", fmt.Errorf("runner is not installed; run `runner install` from the extracted release bundle")
}

func Validate(dir string) error {
	for _, name := range required {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("missing asset %q in %s: %w", name, dir, err)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, RootfsImage)); err != nil {
		return fmt.Errorf("missing asset %q in %s: %w", RootfsImage, dir, err)
	}

	return nil
}

func ValidateBundle(dir string) error {
	for _, name := range required {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("missing asset %q in %s: %w", name, dir, err)
		}
	}

	for _, name := range []string{RootfsImage, RootfsImageCompressed} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return nil
		}
	}

	return fmt.Errorf(
		"missing asset %q or %q in %s",
		RootfsImage,
		RootfsImageCompressed,
		dir,
	)
}

func Symlink(srcDir, vmDir, name string) error {
	target := filepath.Join(srcDir, name)
	link := filepath.Join(vmDir, name)
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("link %s: %w", name, err)
	}
	return nil
}
