package application

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"starliner.app/runner/internal/domain/port"
)

type RunApplication struct {
	starliner port.Starliner
}

func NewRunApplication(starliner port.Starliner) *RunApplication {
	return &RunApplication{
		starliner: starliner,
	}
}

func (a *RunApplication) Start(insecureSkipTLSVerify bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return a.starliner.ServeHeartbeats(ctx, insecureSkipTLSVerify)
}
