package dto

import (
	"fmt"
	"time"

	"starliner.app/runner/internal/domain/value"
)

type RecordDTO struct {
	ID             string `json:"id"`
	Dir            string `json:"dir"`
	Tap            string `json:"tap"`
	MAC            string `json:"mac"`
	SubnetOctet    int    `json:"subnet_octet"`
	GuestCID       uint32 `json:"guest_cid"`
	FirecrackerPID int    `json:"firecracker_pid"`
	DNSMasqPID     int    `json:"dnsmasq_pid"`
	CreatedAt      string `json:"created_at"`
}

func ToDTO(vm value.VM) (RecordDTO, error) {
	return RecordDTO{
		ID:             vm.ID,
		Dir:            vm.Dir,
		Tap:            vm.Tap,
		MAC:            vm.MAC,
		SubnetOctet:    vm.SubnetOctet,
		GuestCID:       vm.GuestCID,
		FirecrackerPID: vm.FirecrackerPID,
		DNSMasqPID:     vm.DNSMasqPID,
		CreatedAt:      vm.CreatedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func FromDTO(dto RecordDTO) (value.VM, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, dto.CreatedAt)
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339, dto.CreatedAt)
		if err != nil {
			return value.VM{}, fmt.Errorf("parse created_at for vm %s: %w", dto.ID, err)
		}
	}

	return value.VM{
		ID:             dto.ID,
		Dir:            dto.Dir,
		Tap:            dto.Tap,
		MAC:            dto.MAC,
		SubnetOctet:    dto.SubnetOctet,
		GuestCID:       dto.GuestCID,
		FirecrackerPID: dto.FirecrackerPID,
		DNSMasqPID:     dto.DNSMasqPID,
		CreatedAt:      createdAt,
	}, nil
}
