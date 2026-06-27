package network

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"starliner.app/runner/internal/infrastructure/privileged"
)

type Host struct {
	tap        string
	egress     string
	createdTap bool
}

func Setup(tap string, subnetOctet int) (*Host, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("network setup requires linux")
	}

	hostCIDR := fmt.Sprintf("172.16.%d.1/24", subnetOctet)

	setup := &Host{tap: tap}

	if !linkExists(tap) {
		if err := privileged.Run("ip", "tuntap", "add", "dev", tap, "mode", "tap"); err != nil {
			return nil, fmt.Errorf("create %s: %w", tap, err)
		}
		setup.createdTap = true
	}

	if err := privileged.Run("ip", "addr", "replace", hostCIDR, "dev", tap); err != nil {
		setup.Teardown()
		return nil, fmt.Errorf("configure %s address: %w", tap, err)
	}

	if err := privileged.Run("ip", "link", "set", "dev", tap, "up"); err != nil {
		setup.Teardown()
		return nil, fmt.Errorf("bring up %s: %w", tap, err)
	}

	if err := privileged.RunQuiet("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		setup.Teardown()
		return nil, fmt.Errorf("enable ip forwarding: %w", err)
	}

	egress, err := defaultEgressInterface()
	if err != nil {
		setup.Teardown()
		return nil, err
	}
	setup.egress = egress

	if err := ensureNAT(egress); err != nil {
		setup.Teardown()
		return nil, err
	}

	if err := ensureIptablesRule(
		[]string{"-C", "FORWARD", "-i", tap, "-o", egress, "-j", "ACCEPT"},
		[]string{"-A", "FORWARD", "-i", tap, "-o", egress, "-j", "ACCEPT"},
	); err != nil {
		setup.Teardown()
		return nil, fmt.Errorf("configure forward rule: %w", err)
	}

	if err := ensureIptablesRule(
		[]string{
			"-C", "FORWARD", "-i", egress, "-o", tap,
			"-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT",
		},
		[]string{
			"-A", "FORWARD", "-i", egress, "-o", tap,
			"-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT",
		},
	); err != nil {
		setup.Teardown()
		return nil, fmt.Errorf("configure return traffic rule: %w", err)
	}

	return setup, nil
}

func ensureNAT(egress string) error {
	if err := ensureIptablesRule(
		[]string{"-t", "nat", "-C", "POSTROUTING", "-o", egress, "-j", "MASQUERADE"},
		[]string{"-t", "nat", "-A", "POSTROUTING", "-o", egress, "-j", "MASQUERADE"},
	); err != nil {
		return fmt.Errorf("configure NAT: %w", err)
	}
	return nil
}

func ensureIptablesRule(checkArgs, addArgs []string) error {
	if hasIptablesRule(checkArgs...) {
		return nil
	}
	return privileged.Run("iptables", addArgs...)
}

func hasIptablesRule(args ...string) bool {
	cmd := privileged.Command("iptables", args...)
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func (n *Host) Teardown() {
	if n.createdTap {
		_ = privileged.Run("ip", "link", "del", n.tap)
	}
}

func Destroy(tap string) {
	if linkExists(tap) {
		_ = privileged.Run("ip", "link", "del", tap)
	}
}

func linkExists(name string) bool {
	cmd := exec.Command("ip", "link", "show", name)
	return cmd.Run() == nil
}

func defaultEgressInterface() (string, error) {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", fmt.Errorf("find default route: %w", err)
	}

	fields := strings.Fields(string(out))
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			return fields[i+1], nil
		}
	}

	return "", fmt.Errorf("no default route found")
}
