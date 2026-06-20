package workload

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"workload",
	fx.Provide(
		fx.Annotate(
			NewTracker,
			fx.As(new(port.WorkloadTracker)),
		),
	),
)
