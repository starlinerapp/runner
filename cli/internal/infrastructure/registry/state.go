package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
	"starliner.app/runner/internal/infrastructure/registry/dto"
)

const registryFile = "vms.json"

type Store struct {
	mu sync.Mutex
}

type fileRegistry struct {
	VMs []dto.RecordDTO `json:"vms"`
}

type mutable struct {
	vms []value.VM
}

func New() port.VMRegistry {
	return &Store{}
}

func (s *Store) List() ([]value.VM, error) {
	reg, err := load()
	if err != nil {
		return nil, err
	}
	return reg.vms, nil
}

func (s *Store) Get(id string) (*value.VM, error) {
	reg, err := load()
	if err != nil {
		return nil, err
	}

	for _, vm := range reg.vms {
		if vm.ID == id {
			return new(vm), nil
		}
	}

	return nil, fmt.Errorf("vm %s not found", id)
}

func (s *Store) WithLock(fn func(port.MutableVMRegistry) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reg, err := load()
	if err != nil {
		return err
	}

	m := &mutable{vms: reg.vms}
	if err := fn(m); err != nil {
		return err
	}

	return save(m.vms)
}

func (m *mutable) VMs() []value.VM {
	return m.vms
}

func (m *mutable) Add(v value.VM) {
	m.vms = append(m.vms, v)
}

func (m *mutable) Remove(id string) (value.VM, bool) {
	for i, vm := range m.vms {
		if vm.ID == id {
			m.vms = append(m.vms[:i], m.vms[i+1:]...)
			return vm, true
		}
	}
	return value.VM{}, false
}

func load() (*mutable, error) {
	path, err := registryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &mutable{}, nil
		}
		return nil, fmt.Errorf("read vm registry: %w", err)
	}

	var reg fileRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse vm registry: %w", err)
	}

	vms := make([]value.VM, 0, len(reg.VMs))
	for _, d := range reg.VMs {
		vm, err := dto.FromDTO(d)
		if err != nil {
			return nil, err
		}
		vms = append(vms, vm)
	}

	return &mutable{vms: vms}, nil
}

func save(vms []value.VM) error {
	path, err := registryPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	dtos := make([]dto.RecordDTO, len(vms))
	for i, vm := range vms {
		d, err := dto.ToDTO(vm)
		if err != nil {
			return err
		}
		dtos[i] = d
	}

	data, err := json.MarshalIndent(fileRegistry{VMs: dtos}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal vm registry: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write vm registry: %w", err)
	}

	return nil
}

func registryPath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, registryFile), nil
}
