package handler

import "go.uber.org/fx"

var Module = fx.Module(
	"handler",
	fx.Provide(
		NewInstallHandler,
		NewVMHandler,
		NewBuildHandler,
	),
)
