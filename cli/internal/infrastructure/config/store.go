package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/infrastructure/firecracker/assets"
	"starliner.app/runner/internal/infrastructure/privileged"
)

const configFile = "config.json"

type fileConfig struct {
	BaseURL string `json:"base_url"`
}

type Store struct{}

func New() port.ConfigStore {
	return &Store{}
}

func (s *Store) SaveBaseURL(baseURL string) error {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "runner-config-*")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	data, err := json.MarshalIndent(fileConfig{BaseURL: normalized}, "", "  ")
	if err != nil {
		cleanup()
		return fmt.Errorf("marshal config: %w", err)
	}

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		cleanup()
		return fmt.Errorf("write temp config: %w", err)
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	dst := filepath.Join(assets.InstalledAssetsDir, configFile)
	if err := privilegedInstall(tmpPath, dst, 0o644); err != nil {
		cleanup()
		return fmt.Errorf("install config: %w", err)
	}

	cleanup()
	return nil
}

func (s *Store) BaseURL() (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("runner config not found; reinstall with --base-url")
		}
		return "", fmt.Errorf("read config: %w", err)
	}

	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse config: %w", err)
	}

	return normalizeBaseURL(cfg.BaseURL)
}

func configPath() (string, error) {
	dir, err := assets.ResolveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("base URL is required")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("base URL must use http or https")
	}
	if u.Host == "" {
		return "", fmt.Errorf("base URL must include a host")
	}

	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}

func privilegedInstall(src, dst string, mode os.FileMode) error {
	if err := os.Chmod(src, mode); err != nil {
		return err
	}
	return privileged.Run("install", "-m", fmt.Sprintf("%o", mode), src, dst)
}
