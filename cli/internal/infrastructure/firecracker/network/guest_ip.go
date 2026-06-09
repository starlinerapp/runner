package network

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func GuestIP(subnetOctet int) string {
	return fmt.Sprintf("172.16.%d.2", subnetOctet)
}

func WaitGuestIP(tap, mac string, timeout time.Duration) (string, error) {
	mac = strings.ToLower(mac)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ip, ok, err := neighborIP(tap, mac)
		if err != nil {
			return "", err
		}
		if ok {
			return ip, nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return "", fmt.Errorf("timed out waiting for guest %s on %s", mac, tap)
}

func neighborIP(tap, mac string) (string, bool, error) {
	out, err := exec.Command("ip", "neigh", "show", "dev", tap).Output()
	if err != nil {
		return "", false, fmt.Errorf("list neighbors on %s: %w", tap, err)
	}

	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		ip := fields[0]
		if fields[1] != "lladdr" || !strings.EqualFold(fields[2], mac) {
			continue
		}

		state := fields[len(fields)-1]
		if state == "FAILED" || state == "INCOMPLETE" {
			continue
		}

		return ip, true, nil
	}

	return "", false, nil
}
