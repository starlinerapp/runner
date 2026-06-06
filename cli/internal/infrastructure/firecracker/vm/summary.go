package vm

import (
	"fmt"

	domainvm "starliner.app/runner/internal/domain/value"
)

func PrintSummary(rec *domainvm.VM) {
	fmt.Printf("VM %s created\n", rec.ID)
	fmt.Printf("  tap:         %s\n", rec.Tap)
	fmt.Printf("  mac:         %s\n", rec.MAC)
	fmt.Printf("  subnet:      172.16.%d.0/24\n", rec.SubnetOctet)
	fmt.Printf("  guest cid:   %d\n", rec.GuestCID)
	fmt.Printf("  workspace:   %s\n", rec.Dir)
	fmt.Printf("  firecracker: pid %d\n", rec.FirecrackerPID)
}
