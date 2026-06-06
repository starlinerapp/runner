package vm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"starliner.app/runner/internal/infrastructure/firecracker/allocate"
	"starliner.app/runner/internal/infrastructure/firecracker/assets"
	"starliner.app/runner/internal/infrastructure/firecracker/config"
	"starliner.app/runner/internal/infrastructure/registry"
)

func Start(vmDir string) (*exec.Cmd, error) {
	configPath := filepath.Join(vmDir, config.FileName)
	socketPath := config.SocketPath(vmDir)

	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("remove stale api socket: %w", err)
	}

	cmd := exec.Command("firecracker", "--api-sock", socketPath, "--config-file", configPath)
	cmd.Dir = vmDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start firecracker: %w", err)
	}

	if err := waitRunning(cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}

func waitRunning(cmd *exec.Cmd) error {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("firecracker exited: %w", err)
		}
		return fmt.Errorf("firecracker exited immediately")
	case <-time.After(200 * time.Millisecond):
		return nil
	}
}

func copyRootfs(src, dst string) error {
	if _, err := exec.LookPath("cp"); err == nil {
		cmd := exec.Command("cp", "--reflink=auto", src, dst)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open rootfs template: %w", err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create vm rootfs: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy rootfs: %w", err)
	}

	return nil
}

func prepareWorkspace(assetsDir string, res allocate.Resources) (string, error) {
	dir, err := registry.VMDir(res.ID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create vm dir: %w", err)
	}

	rootfsSrc := filepath.Join(assetsDir, assets.RootfsImage)
	rootfsDst := filepath.Join(dir, assets.RootfsImage)
	if err := copyRootfs(rootfsSrc, rootfsDst); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}

	for _, name := range []string{assets.KernelImage, assets.InitrdImage} {
		if err := assets.Symlink(assetsDir, dir, name); err != nil {
			_ = os.RemoveAll(dir)
			return "", err
		}
	}

	configPath := filepath.Join(dir, config.FileName)
	if err := config.Write(configPath, res.Tap, res.MAC); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}
