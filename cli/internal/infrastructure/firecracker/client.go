package firecracker

import (
	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
	"starliner.app/runner/internal/infrastructure/firecracker/assets"
	"starliner.app/runner/internal/infrastructure/firecracker/vm"
)

type Client struct {
	registry port.VMRegistry
}

func NewClient(registry port.VMRegistry) port.VM {
	return &Client{registry: registry}
}

func (c *Client) Create() error {
	assetsDir, err := assets.ResolveDir()
	if err != nil {
		return err
	}

	if err := assets.Validate(assetsDir); err != nil {
		return err
	}

	record, err := vm.Create(c.registry, assetsDir)
	if err != nil {
		return err
	}

	vm.PrintSummary(record)
	return nil
}

func (c *Client) List() ([]value.VM, error) {
	return c.registry.List()
}

func (c *Client) Delete(id string) error {
	return vm.Delete(c.registry, id)
}
