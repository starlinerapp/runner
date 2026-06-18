package starliner

import "go.uber.org/fx"

var Module = fx.Module(
	"starliner",
	fx.Provide(
		NewClient,
	),
)
