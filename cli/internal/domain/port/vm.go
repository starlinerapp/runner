package port

import "starliner.app/runner/internal/domain/value"

type VM interface {
	Create() error
	List() ([]value.VM, error)
	Delete(id string) error
}
