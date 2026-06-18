package starliner

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"starliner",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.As(new(port.Starliner)),
		),
	),
)
