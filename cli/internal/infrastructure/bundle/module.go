package bundle

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"bundle",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.As(new(port.Installer)),
		),
	),
)
