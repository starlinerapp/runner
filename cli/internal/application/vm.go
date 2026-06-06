package application

import (
	"starliner.app/runner/internal/domain/port"
)

type VMApplication struct {
	VM port.VM
}

func NewVMApplication(
	VM port.VM,
) *VMApplication {
	return &VMApplication{
		VM: VM,
	}
}

func (vm *VMApplication) CreateVM() error {
	return vm.VM.Create()
}
