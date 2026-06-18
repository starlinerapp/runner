package credentials

import "go.uber.org/fx"

var Module = fx.Module(
	"credentials",
	fx.Provide(New),
)
