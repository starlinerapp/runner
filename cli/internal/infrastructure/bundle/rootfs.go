package bundle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"starliner.app/runner/internal/infrastructure/firecracker/assets"
)

func writeBundleRootfs(dest, bundleDir string) error {
	uncompressed := filepath.Join(bundleDir, assets.RootfsImage)
	if _, err := os.Stat(uncompressed); err == nil {
		return copyRootfsFile(uncompressed, dest)
	}

	compressed := filepath.Join(bundleDir, assets.RootfsImageCompressed)
	if _, err := os.Stat(compressed); err != nil {
		return fmt.Errorf("rootfs asset: %w", err)
	}

	return decompressRootfs(compressed, dest)
}

func copyRootfsFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open rootfs: %w", err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create rootfs: %w", err)
	}

	cleanup := func() {
		_ = out.Close()
		_ = os.Remove(dest)
	}

	if _, err := io.Copy(out, in); err != nil {
		cleanup()
		return fmt.Errorf("copy rootfs: %w", err)
	}

	if err := out.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close rootfs: %w", err)
	}

	return nil
}

func decompressRootfs(compressed, dest string) error {
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

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create rootfs: %w", err)
	}

	cleanup := func() {
		_ = out.Close()
		_ = os.Remove(dest)
	}

	if _, err := io.Copy(out, decoder); err != nil {
		cleanup()
		return fmt.Errorf("decompress rootfs: %w", err)
	}

	if err := out.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close rootfs: %w", err)
	}

	return nil
}
