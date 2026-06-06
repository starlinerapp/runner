package firecracker

import (
	"os"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
	"starliner.app/runner/internal/infrastructure/firecracker/assets"
	"starliner.app/runner/internal/infrastructure/firecracker/network"
	"starliner.app/runner/internal/infrastructure/firecracker/vm"
)

type Runtime struct{}

func NewRuntime() port.VMRuntime {
	return &Runtime{}
}

func (r *Runtime) Provision(res value.VMResources) (*port.ProvisionResult, error) {
	assetsDir, err := assets.ResolveDir()
	if err != nil {
		return nil, err
	}

	if err := assets.Validate(assetsDir); err != nil {
		return nil, err
	}

	dir, err := vm.PrepareWorkspace(assetsDir, res)
	if err != nil {
		return nil, err
	}

	net, err := network.Setup(res.Tap, res.SubnetOctet)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}

	vmCmd, err := vm.Start(dir)
	if err != nil {
		net.Teardown()
		_ = os.RemoveAll(dir)
		return nil, err
	}

	return &port.ProvisionResult{
		Dir:            dir,
		FirecrackerPID: vmCmd.Process.Pid,
		DNSMasqPID:     net.DNSMasqPID(),
	}, nil
}

func (r *Runtime) Teardown(record value.VM) error {
	vm.Teardown(record)
	return nil
}
