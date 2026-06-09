package network

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"starliner.app/runner/internal/infrastructure/privileged"
)

type Host struct {
	tap        string
	dnsmasq    *exec.Cmd
	egress     string
	createdTap bool
}

func Setup(tap string, subnetOctet int, mac string) (*Host, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("network setup requires linux")
	}

	hostCIDR := fmt.Sprintf("172.16.%d.1/24", subnetOctet)
	dhcpRange := fmt.Sprintf("172.16.%d.2,172.16.%d.254,255.255.255.0", subnetOctet, subnetOctet)
	gateway := fmt.Sprintf("172.16.%d.1", subnetOctet)

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

	if err := privileged.Run("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
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

	dnsmasq, err := startDHCP(tap, dhcpRange, gateway, mac, GuestIP(subnetOctet))
	if err != nil {
		setup.Teardown()
		return nil, err
	}
	setup.dnsmasq = dnsmasq

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

func (n *Host) DNSMasqPID() int {
	if n.dnsmasq != nil && n.dnsmasq.Process != nil {
		return n.dnsmasq.Process.Pid
	}
	return 0
}

func (n *Host) Teardown() {
	if n.dnsmasq != nil && n.dnsmasq.Process != nil {
		_ = n.dnsmasq.Process.Kill()
	}

	if n.createdTap {
		_ = privileged.Run("ip", "link", "del", n.tap)
	}
}

func Destroy(tap string, dnsmasqPID int) {
	if dnsmasqPID > 0 {
		if proc, err := os.FindProcess(dnsmasqPID); err == nil {
			_ = proc.Kill()
		}
	}

	if linkExists(tap) {
		_ = privileged.Run("ip", "link", "del", tap)
	}
}

func startDHCP(tap, dhcpRange, gateway, mac, guestIP string) (*exec.Cmd, error) {
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		return nil, fmt.Errorf("dnsmasq not found (required for guest DHCP): %w", err)
	}

	cmd := privileged.Command(
		"dnsmasq",
		"--keep-in-foreground",
		"--bind-interfaces",
		"--interface="+tap,
		"--dhcp-range="+dhcpRange,
		"--dhcp-host="+mac+","+guestIP,
		"--dhcp-option=option:router,"+gateway,
		"--dhcp-option=option:dns-server,8.8.8.8",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start dnsmasq: %w", err)
	}

	return cmd, nil
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
