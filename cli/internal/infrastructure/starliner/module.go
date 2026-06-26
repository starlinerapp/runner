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
		fx.Annotate(
			NewJobGateway,
			fx.As(new(port.JobGateway)),
		),
		fx.Annotate(
			NewHeartbeatGateway,
			fx.As(new(port.HeartbeatGateway)),
		),
	),
)
