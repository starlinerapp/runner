package value

import "time"

type VM struct {
	ID             string
	Dir            string
	Tap            string
	MAC            string
	GuestIP        string
	SubnetOctet    int
	GuestCID       uint32
	FirecrackerPID int
	DNSMasqPID     int
	CreatedAt      time.Time
}
