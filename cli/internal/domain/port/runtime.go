package port

import "starliner.app/runner/internal/domain/value"

type ProvisionResult struct {
	Dir            string
	GuestIP        string
	FirecrackerPID int
	DNSMasqPID     int
}

type VMRuntime interface {
	Provision(res value.VMResources) (*ProvisionResult, error)
	Teardown(vm value.VM) error
}
