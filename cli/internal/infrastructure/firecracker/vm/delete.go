package vm

import (
	"fmt"
	"os"
	"syscall"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
	"starliner.app/runner/internal/infrastructure/firecracker/network"
)

func Delete(reg port.VMRegistry, id string) error {
	record, err := reg.Get(id)
	if err != nil {
		return err
	}

	teardown(*record)

	return reg.WithLock(func(m port.MutableVMRegistry) error {
		if _, ok := m.Remove(id); !ok {
			return fmt.Errorf("vm %s not found", id)
		}
		return nil
	})
}

func teardown(record value.VM) {
	stopProcess(record.FirecrackerPID)
	network.Destroy(record.Tap, record.DNSMasqPID)
	_ = os.RemoveAll(record.Dir)
}

func stopProcess(pid int) {
	if pid <= 0 {
		return
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	_ = proc.Signal(syscall.SIGTERM)
}
