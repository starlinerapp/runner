package firecracker

import "go.uber.org/fx"

var Module = fx.Module(
	"firecracker",
	fx.Provide(NewRuntime),
)
