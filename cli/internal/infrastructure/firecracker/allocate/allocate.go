package allocate

import (
	"crypto/rand"
	"fmt"

	domainvm "starliner.app/runner/internal/domain/value"
)

const (
	minGuestCID    = uint32(3)
	maxSubnetOctet = 254
)

type Resources struct {
	ID          string
	Tap         string
	MAC         string
	SubnetOctet int
	GuestCID    uint32
}

func ForFleet(vms []domainvm.VM) (Resources, error) {
	id, err := randomVMID()
	if err != nil {
		return Resources{}, err
	}

	mac, err := randomMAC()
	if err != nil {
		return Resources{}, err
	}

	subnetOctet, err := nextSubnetOctet(vms)
	if err != nil {
		return Resources{}, err
	}

	guestCID, err := nextGuestCID(vms)
	if err != nil {
		return Resources{}, err
	}

	return Resources{
		ID:          id,
		Tap:         fmt.Sprintf("rtap-%s", id),
		MAC:         mac,
		SubnetOctet: subnetOctet,
		GuestCID:    guestCID,
	}, nil
}

func randomVMID() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate vm id: %w", err)
	}
	return fmt.Sprintf("%x", buf), nil
}

func randomMAC() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate mac: %w", err)
	}
	return fmt.Sprintf("AA:FC:%02X:%02X:%02X:%02X", buf[0], buf[1], buf[2], buf[3]), nil
}

func nextSubnetOctet(vms []domainvm.VM) (int, error) {
	used := make(map[int]struct{}, len(vms))
	for _, vm := range vms {
		used[vm.SubnetOctet] = struct{}{}
	}

	for octet := 1; octet <= maxSubnetOctet; octet++ {
		if _, ok := used[octet]; !ok {
			return octet, nil
		}
	}

	return 0, fmt.Errorf("no free subnet octets (max %d concurrent VMs)", maxSubnetOctet)
}

func nextGuestCID(vms []domainvm.VM) (uint32, error) {
	used := make(map[uint32]struct{}, len(vms))
	for _, vm := range vms {
		used[vm.GuestCID] = struct{}{}
	}

	for cid := minGuestCID; cid <= minGuestCID+uint32(maxSubnetOctet); cid++ {
		if _, ok := used[cid]; !ok {
			return cid, nil
		}
	}

	return 0, fmt.Errorf("no free guest CIDs")
}
