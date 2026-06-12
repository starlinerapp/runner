package buildkit

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	Port           = 1234
	ConnectTimeout = 5 * time.Minute
)

func Wait(host string, timeout time.Duration) error {
	addr := net.JoinHostPort(host, strconv.Itoa(Port))
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if tcpOpen(addr, 500*time.Millisecond) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timed out waiting for buildkit at %s", addr)
}

func tcpOpen(addr string, dialTimeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
