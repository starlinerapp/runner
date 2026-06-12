package config

import (
	"fmt"
	"strings"
)

func AppendGuestNetwork(bootArgs string, subnetOctet int) string {
	guestIP := fmt.Sprintf("172.16.%d.2", subnetOctet)
	gateway := fmt.Sprintf("172.16.%d.1", subnetOctet)
	netArgs := fmt.Sprintf("runner.ipv4=%s/24 runner.ipv4_gw=%s", guestIP, gateway)
	return strings.TrimSpace(bootArgs) + " " + netArgs
}
