package bundle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/infrastructure/firecracker/assets"
	"starliner.app/runner/internal/infrastructure/firecracker/install"
	"starliner.app/runner/internal/infrastructure/privileged"
)

const BinaryPath = assets.InstalledBinaryDir + "/runner"

type Client struct{}

func NewClient() port.Installer {
	return &Client{}
}

func (c *Client) Install() error {
	sourceDir, err := sourceDir()
	if err != nil {
		return err
	}

	if err := assets.ValidateBundle(sourceDir); err != nil {
		return fmt.Errorf("bundle assets: %w", err)
	}

	if err := privileged.Run("mkdir", "-p", assets.InstalledAssetsDir); err != nil {
		return fmt.Errorf("create assets dir: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve runner binary: %w", err)
	}

	if err := installFile(exe, BinaryPath, 0o755); err != nil {
		return fmt.Errorf("install runner binary: %w", err)
	}

	for _, name := range assetFiles() {
		src := filepath.Join(sourceDir, name)
		dst := filepath.Join(assets.InstalledAssetsDir, name)
		if err := installFile(src, dst, 0o644); err != nil {
			return fmt.Errorf("install asset %q: %w", name, err)
		}
	}

	if err := installRootfs(sourceDir); err != nil {
		return fmt.Errorf("install rootfs: %w", err)
	}

	if err := install.Ensure(); err != nil {
		return err
	}

	fmt.Println("Runner installed successfully")
	fmt.Printf("  binary: %s\n", BinaryPath)
	fmt.Printf("  assets: %s\n", assets.InstalledAssetsDir)
	return nil
}

func sourceDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve bundle dir: %w", err)
	}

	dir, err := filepath.EvalSymlinks(filepath.Dir(exe))
	if err != nil {
		return "", fmt.Errorf("resolve bundle dir: %w", err)
	}

	if dir == assets.InstalledBinaryDir {
		return "", fmt.Errorf("run install from the extracted release bundle, not the installed binary")
	}

	return dir, nil
}

func assetFiles() []string {
	return []string{
		assets.KernelImage,
		assets.InitrdImage,
		assets.BootArgsFile,
	}
}

func installRootfs(sourceDir string) error {
	tmp, err := os.CreateTemp("", assets.RootfsImage+".*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()

	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	fmt.Println("Decompressing rootfs...")
	if err := writeBundleRootfs(tmpPath, sourceDir); err != nil {
		cleanup()
		return err
	}

	dst := filepath.Join(assets.InstalledAssetsDir, assets.RootfsImage)
	if err := privilegedInstall(tmpPath, dst, 0o644); err != nil {
		cleanup()
		return err
	}

	cleanup()
	return nil
}

func installFile(src, dst string, mode os.FileMode) error {
	tmp, err := os.CreateTemp("", "runner-install-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	in, err := os.Open(src)
	if err != nil {
		cleanup()
		return err
	}
	defer func() { _ = in.Close() }()

	if _, err := io.Copy(tmp, in); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	if err := privilegedInstall(tmpPath, dst, mode); err != nil {
		cleanup()
		return err
	}

	_ = os.Remove(tmpPath)
	return nil
}

func privilegedInstall(src, dst string, mode os.FileMode) error {
	if err := os.Chmod(src, mode); err != nil {
		return err
	}
	return privileged.Run("install", "-m", fmt.Sprintf("%o", mode), src, dst)
}
