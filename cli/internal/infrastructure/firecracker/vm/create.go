package vm

import (
	"os"
	"time"

	"starliner.app/runner/internal/infrastructure/firecracker/network"
	"starliner.app/runner/internal/infrastructure/firecracker/registry"
)

func Create(assetsDir string) (*registry.Record, error) {
	var record *registry.Record

	err := registry.WithLock(func(reg *registry.Registry) error {
		res, err := registry.Allocate(reg)
		if err != nil {
			return err
		}

		dir, err := prepareWorkspace(assetsDir, res)
		if err != nil {
			return err
		}

		net, err := network.Setup(res.Tap, res.SubnetOctet)
		if err != nil {
			_ = os.RemoveAll(dir)
			return err
		}

		vmCmd, err := Start(dir)
		if err != nil {
			net.Teardown()
			_ = os.RemoveAll(dir)
			return err
		}

		rec := registry.Record{
			ID:             res.ID,
			Dir:            dir,
			Tap:            res.Tap,
			MAC:            res.MAC,
			SubnetOctet:    res.SubnetOctet,
			GuestCID:       res.GuestCID,
			FirecrackerPID: vmCmd.Process.Pid,
			DNSMasqPID:     net.DNSMasqPID(),
			CreatedAt:      time.Now().UTC(),
		}

		reg.VMs = append(reg.VMs, rec)
		record = &rec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return record, nil
}
