package application

import (
	"fmt"
	"time"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/service/fleet"
	"starliner.app/runner/internal/domain/value"
)

type VMApplication struct {
	registry port.VMRegistry
	runtime  port.VMRuntime
}

func NewVMApplication(
	registry port.VMRegistry,
	runtime port.VMRuntime,
) *VMApplication {
	return &VMApplication{
		registry: registry,
		runtime:  runtime,
	}
}

func (a *VMApplication) CreateVM() (*value.VM, error) {
	var record *value.VM

	err := a.registry.WithLock(func(m port.MutableVMRegistry) error {
		res, err := fleet.Allocate(m.VMs())
		if err != nil {
			return err
		}

		provisioned, err := a.runtime.Provision(res)
		if err != nil {
			return err
		}

		rec := value.VM{
			ID:             res.ID,
			Dir:            provisioned.Dir,
			Tap:            res.Tap,
			MAC:            res.MAC,
			GuestIP:        provisioned.GuestIP,
			SubnetOctet:    res.SubnetOctet,
			GuestCID:       res.GuestCID,
			FirecrackerPID: provisioned.FirecrackerPID,
			DNSMasqPID:     provisioned.DNSMasqPID,
			CreatedAt:      time.Now().UTC(),
		}

		m.Add(rec)
		record = &rec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (a *VMApplication) ListVMs() ([]value.VM, error) {
	return a.registry.List()
}

func (a *VMApplication) Diagnose(vm value.VM) {
	a.runtime.Diagnose(vm)
}

func (a *VMApplication) DeleteVM(id string) error {
	record, err := a.registry.Get(id)
	if err != nil {
		return err
	}

	err = a.runtime.Teardown(*record)
	if err != nil {
		return err
	}

	return a.registry.WithLock(func(m port.MutableVMRegistry) error {
		if _, ok := m.Remove(id); !ok {
			return fmt.Errorf("vm %s not found", id)
		}
		return nil
	})
}
