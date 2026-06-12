package port

import "starliner.app/runner/internal/domain/value"

type VM interface {
	CreateVM() (*value.VM, error)
	DeleteVM(id string) error
	Diagnose(vm value.VM)
}
