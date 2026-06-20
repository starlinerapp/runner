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
	Token             string `json:"token"`
	MaxConcurrentJobs int    `json:"maxConcurrentJobs"`
}

type Store struct{}

func New() port.RegistrationStore {
	return &Store{}
}

func (s *Store) SaveRegistration(registration port.RunnerRegistration) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create credentials dir: %w", err)
	}

	data, err := json.MarshalIndent(fileCredentials{
		Token:             registration.Token,
		MaxConcurrentJobs: registration.MaxConcurrentJobs,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

func (s *Store) Token() (string, error) {
	creds, err := s.readCredentials()
	if err != nil {
		return "", err
	}

	return creds.Token, nil
}

func (s *Store) MaxConcurrentJobs() (int, error) {
	creds, err := s.readCredentials()
	if err != nil {
		return 0, err
	}
	if creds.MaxConcurrentJobs < 1 {
		return 0, fmt.Errorf("max concurrent jobs not configured; run `runner register`")
	}

	return creds.MaxConcurrentJobs, nil
}

func (s *Store) readCredentials() (*fileCredentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileCredentials{}, nil
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var creds fileCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	return &creds, nil
}

func credentialsPath() (string, error) {
	dir, err := registry.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}
