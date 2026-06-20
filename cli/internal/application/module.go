package application

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"application",
	fx.Provide(
		NewInstallApplication,
		NewVMApplication,
		NewBuildApplication,
		NewRegisterApplication,
		NewWorkerApplication,
		NewRunApplication,
		fx.Annotate(
			func(a *VMApplication) port.VM { return a },
			fx.As(new(port.VM)),
		),
		fx.Annotate(
			func(a *BuildApplication) port.JobExecutor { return a },
			fx.As(new(port.JobExecutor)),
		),
	),
)
