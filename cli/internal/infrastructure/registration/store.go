package registration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/infrastructure/registry"
)

const registrationFile = "registration.json"

type fileRegistration struct {
	Token             string `json:"token"`
	Name              string `json:"name"`
	MaxConcurrentJobs int    `json:"maxConcurrentJobs"`
}

type Store struct{}

func New() port.RegistrationStore {
	return &Store{}
}

func (s *Store) SaveRegistration(registration port.RunnerRegistration) error {
	path, err := registrationPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create registration dir: %w", err)
	}

	data, err := json.MarshalIndent(fileRegistration{
		Token:             registration.Token,
		Name:              registration.Name,
		MaxConcurrentJobs: registration.MaxConcurrentJobs,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registration: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write registration: %w", err)
	}

	return nil
}

func (s *Store) Token() (string, error) {
	reg, err := s.readRegistration()
	if err != nil {
		return "", err
	}

	return reg.Token, nil
}

func (s *Store) Name() (string, error) {
	reg, err := s.readRegistration()
	if err != nil {
		return "", err
	}

	return reg.Name, nil
}

func (s *Store) MaxConcurrentJobs() (int, error) {
	reg, err := s.readRegistration()
	if err != nil {
		return 0, err
	}
	if reg.MaxConcurrentJobs < 1 {
		return 0, fmt.Errorf("max concurrent jobs not configured; run `runner register`")
	}

	return reg.MaxConcurrentJobs, nil
}

func (s *Store) readRegistration() (*fileRegistration, error) {
	path, err := registrationPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileRegistration{}, nil
		}
		return nil, fmt.Errorf("read registration: %w", err)
	}

	var reg fileRegistration
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registration: %w", err)
	}

	return &reg, nil
}

func registrationPath() (string, error) {
	dir, err := registry.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, registrationFile), nil
}
