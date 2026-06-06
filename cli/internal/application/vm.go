package application

import (
	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
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

func (vm *VMApplication) ListVMs() ([]value.VM, error) {
	return vm.VM.List()
}

func (vm *VMApplication) DeleteVM(id string) error {
	return vm.VM.Delete(id)
}
