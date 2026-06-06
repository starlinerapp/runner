package vm

import (
	"os"
	"time"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
	"starliner.app/runner/internal/infrastructure/firecracker/allocate"
	"starliner.app/runner/internal/infrastructure/firecracker/network"
)

func Create(reg port.VMRegistry, assetsDir string) (*value.VM, error) {
	var record *value.VM

	err := reg.WithLock(func(m port.MutableVMRegistry) error {
		res, err := allocate.ForFleet(m.VMs())
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

		rec := value.VM{
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

		m.Add(rec)
		record = &rec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return record, nil
}
