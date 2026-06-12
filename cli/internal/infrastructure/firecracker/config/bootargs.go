package config

import (
	"fmt"
	"strings"
)

func AppendGuestNetwork(bootArgs string, subnetOctet int) string {
	guestIP := fmt.Sprintf("172.16.%d.2", subnetOctet)
	gateway := fmt.Sprintf("172.16.%d.1", subnetOctet)
	ipArg := fmt.Sprintf("ip=%s::%s:255.255.255.0::eth0:off", guestIP, gateway)
	return strings.TrimSpace(bootArgs) + " " + ipArg
}
