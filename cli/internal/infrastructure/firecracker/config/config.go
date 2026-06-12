package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"starliner.app/runner/internal/infrastructure/firecracker/assets"
)

const (
	FileName      = "firecracker.json"
	SocketName    = "firecracker.socket"
	LogFileName   = "firecracker.log"
	SerialLogName = "serial.log"
	defaultVCPUs  = 2
	defaultMemMiB = 4096
)

type vmConfig struct {
	APISocket struct {
		Path       string `json:"path"`
		UpdatePath bool   `json:"update_path"`
	} `json:"api-socket"`
	SerialConsole *serialConsoleConfig `json:"serial-console,omitempty"`
	BootSource    struct {
		KernelImagePath string `json:"kernel_image_path"`
		InitrdPath      string `json:"initrd_path"`
		BootArgs        string `json:"boot_args"`
	} `json:"boot-source"`
	Drives            []driveConfig            `json:"drives"`
	NetworkInterfaces []networkInterfaceConfig `json:"network-interfaces"`
	MachineConfig     struct {
		VCPUCount  int `json:"vcpu_count"`
		MemSizeMib int `json:"mem_size_mib"`
	} `json:"machine-config"`
}

type driveConfig struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

type networkInterfaceConfig struct {
	IFaceID     string `json:"iface_id"`
	GuestMAC    string `json:"guest_mac"`
	HostDevName string `json:"host_dev_name"`
}

type serialConsoleConfig struct {
	Mode     string `json:"mode"`
	Filename string `json:"filename"`
}

func LogPath(vmDir string) string {
	return filepath.Join(vmDir, LogFileName)
}

func SocketPath(vmDir string) string {
	return filepath.Join(vmDir, SocketName)
}

func Write(destPath, tap, mac, bootArgs string) error {
	vmDir := filepath.Dir(destPath)
	cfg := vmConfig{}
	cfg.APISocket.Path = SocketPath(vmDir)
	cfg.APISocket.UpdatePath = false
	cfg.SerialConsole = &serialConsoleConfig{
		Mode:     "Pipe",
		Filename: "./" + SerialLogName,
	}
	cfg.BootSource.KernelImagePath = "./" + assets.KernelImage
	cfg.BootSource.InitrdPath = "./" + assets.InitrdImage
	cfg.BootSource.BootArgs = bootArgs
	cfg.Drives = []driveConfig{{
		DriveID:      "rootfs",
		PathOnHost:   "./" + assets.RootfsImage,
		IsRootDevice: true,
		IsReadOnly:   false,
	}}
	cfg.NetworkInterfaces = []networkInterfaceConfig{{
		IFaceID:     "eth0",
		GuestMAC:    mac,
		HostDevName: tap,
	}}
	cfg.MachineConfig.VCPUCount = defaultVCPUs
	cfg.MachineConfig.MemSizeMib = defaultMemMiB

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal firecracker config: %w", err)
	}

	if err := os.WriteFile(destPath, append(out, '\n'), 0o644); err != nil {
		return fmt.Errorf("write firecracker config: %w", err)
	}

	return nil
}
