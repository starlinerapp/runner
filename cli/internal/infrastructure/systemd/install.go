package systemd

import (
	_ "embed"
	"fmt"
	"os"

	"starliner.app/runner/internal/infrastructure/privileged"
)

//go:embed starliner-runner.service
var unitFile string

const unitPath = "/etc/systemd/system/starliner-runner.service"

func UnitPath() string {
	return unitPath
}

func Install() error {
	tmp, err := os.CreateTemp("", "starliner-runner.service.*")
	if err != nil {
		return fmt.Errorf("create temp unit file: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.WriteString(unitFile); err != nil {
		cleanup()
		return fmt.Errorf("write temp unit file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	if err := privileged.Run("install", "-m", "0644", tmpPath, unitPath); err != nil {
		cleanup()
		return fmt.Errorf("install systemd unit: %w", err)
	}

	cleanup()

	if err := privileged.Run("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("reload systemd: %w", err)
	}

	return nil
}
