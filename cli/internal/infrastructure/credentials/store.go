package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/infrastructure/registry"
)

const credentialsFile = "credentials.json"

type fileCredentials struct {
	Token string `json:"token"`
}

type Store struct{}

func New() port.CredentialsStore {
	return &Store{}
}

func (s *Store) SaveToken(token string) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create credentials dir: %w", err)
	}

	data, err := json.MarshalIndent(fileCredentials{Token: token}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

func (s *Store) Token() (string, error) {
	path, err := credentialsPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read credentials: %w", err)
	}

	var creds fileCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}

	return creds.Token, nil
}

func credentialsPath() (string, error) {
	dir, err := registry.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}
