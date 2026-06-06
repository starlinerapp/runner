package assets

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
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

	if err := rootfsPresent(dir); err != nil {
		return err
	}

	return nil
}

func rootfsPresent(dir string) error {
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

func EnsureRootfs(dir string) error {
	uncompressed := filepath.Join(dir, RootfsImage)
	if _, err := os.Stat(uncompressed); err == nil {
		return nil
	}

	compressed := filepath.Join(dir, RootfsImageCompressed)
	if _, err := os.Stat(compressed); err != nil {
		return fmt.Errorf("rootfs asset: %w", err)
	}

	in, err := os.Open(compressed)
	if err != nil {
		return fmt.Errorf("open compressed rootfs: %w", err)
	}
	defer func() { _ = in.Close() }()

	decoder, err := zstd.NewReader(in)
	if err != nil {
		return fmt.Errorf("create zstd decoder: %w", err)
	}
	defer decoder.Close()

	tmp, err := os.CreateTemp(dir, RootfsImage+".*")
	if err != nil {
		return fmt.Errorf("create temp rootfs: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := io.Copy(tmp, decoder); err != nil {
		cleanup()
		return fmt.Errorf("decompress rootfs: %w", err)
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp rootfs: %w", err)
	}

	if err := os.Rename(tmpPath, uncompressed); err != nil {
		cleanup()
		return fmt.Errorf("install rootfs: %w", err)
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
