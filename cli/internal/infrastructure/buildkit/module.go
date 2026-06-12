package buildkit

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"buildkit",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.As(new(port.Buildkit)),
		),
	),
)
