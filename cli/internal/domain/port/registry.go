package port

import "starliner.app/runner/internal/domain/value"

type VMRegistry interface {
	List() ([]value.VM, error)
	Get(id string) (*value.VM, error)
	WithLock(fn func(MutableVMRegistry) error) error
}

type MutableVMRegistry interface {
	VMs() []value.VM
	Add(v value.VM)
	Remove(id string) (value.VM, bool)
}
