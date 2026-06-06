package registry

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	minGuestCID    = uint32(3)
	maxSubnetOctet = 254
	registryFile   = "vms.json"
)

type Record struct {
	ID             string    `json:"id"`
	Dir            string    `json:"dir"`
	Tap            string    `json:"tap"`
	MAC            string    `json:"mac"`
	SubnetOctet    int       `json:"subnet_octet"`
	GuestCID       uint32    `json:"guest_cid"`
	FirecrackerPID int       `json:"firecracker_pid"`
	DNSMasqPID     int       `json:"dnsmasq_pid"`
	CreatedAt      time.Time `json:"created_at"`
}

type Resources struct {
	ID          string
	Tap         string
	MAC         string
	SubnetOctet int
	GuestCID    uint32
}

type Registry struct {
	VMs []Record `json:"vms"`
}

var mu sync.Mutex

func StateDir() (string, error) {
	if dir := os.Getenv("RUNNER_STATE_DIR"); dir != "" {
		return dir, nil
	}

	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve state dir: %w", err)
	}

	return filepath.Join(base, "runner"), nil
}

func VMDir(id string) (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vms", id), nil
}

func Load() (*Registry, error) {
	path, err := registryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{}, nil
		}
		return nil, fmt.Errorf("read vm registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse vm registry: %w", err)
	}

	return &reg, nil
}

func Save(reg *Registry) error {
	path, err := registryPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal vm registry: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write vm registry: %w", err)
	}

	return nil
}

func Allocate(reg *Registry) (Resources, error) {
	id, err := randomVMID()
	if err != nil {
		return Resources{}, err
	}

	mac, err := randomMAC()
	if err != nil {
		return Resources{}, err
	}

	subnetOctet, err := nextSubnetOctet(reg)
	if err != nil {
		return Resources{}, err
	}

	guestCID, err := nextGuestCID(reg)
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

func WithLock(fn func(*Registry) error) error {
	mu.Lock()
	defer mu.Unlock()

	reg, err := Load()
	if err != nil {
		return err
	}

	if err := fn(reg); err != nil {
		return err
	}

	return Save(reg)
}

func registryPath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, registryFile), nil
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

func nextSubnetOctet(reg *Registry) (int, error) {
	used := make(map[int]struct{}, len(reg.VMs))
	for _, vm := range reg.VMs {
		used[vm.SubnetOctet] = struct{}{}
	}

	for octet := 1; octet <= maxSubnetOctet; octet++ {
		if _, ok := used[octet]; !ok {
			return octet, nil
		}
	}

	return 0, fmt.Errorf("no free subnet octets (max %d concurrent VMs)", maxSubnetOctet)
}

func nextGuestCID(reg *Registry) (uint32, error) {
	used := make(map[uint32]struct{}, len(reg.VMs))
	for _, vm := range reg.VMs {
		used[vm.GuestCID] = struct{}{}
	}

	for cid := minGuestCID; cid <= minGuestCID+uint32(maxSubnetOctet); cid++ {
		if _, ok := used[cid]; !ok {
			return cid, nil
		}
	}

	return 0, fmt.Errorf("no free guest CIDs")
}
