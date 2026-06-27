package vm

import (
	"os"
	"syscall"

	"starliner.app/runner/internal/domain/value"
	"starliner.app/runner/internal/infrastructure/firecracker/network"
)

func Teardown(record value.VM) {
	stopProcess(record.FirecrackerPID)
	network.Destroy(record.Tap)
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

func ProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}
