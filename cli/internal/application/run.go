package application

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type RunApplication struct {
	workerApplication *WorkerApplication
}

func NewRunApplication(workerApplication *WorkerApplication) *RunApplication {
	return &RunApplication{
		workerApplication: workerApplication,
	}
}

func (a *RunApplication) Start(insecureSkipTLSVerify bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return a.workerApplication.Run(ctx, insecureSkipTLSVerify)
}
