package port

import (
	"context"
)

type HeartbeatSession interface {
	Run(ctx context.Context) error
	Close() error
}

type HeartbeatGateway interface {
	OpenHeartbeatSession(ctx context.Context, insecureSkipTLSVerify bool) (HeartbeatSession, error)
}
