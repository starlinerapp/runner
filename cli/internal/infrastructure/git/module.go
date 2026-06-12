package git

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"git",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.As(new(port.Git)),
		),
	),
)
