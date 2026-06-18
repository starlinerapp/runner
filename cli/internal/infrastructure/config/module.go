package config

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"config",
	fx.Provide(
		fx.Annotate(
			New,
			fx.As(new(port.ConfigStore)),
		),
	),
)
