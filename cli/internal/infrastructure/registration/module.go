package registration

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/domain/port"
)

var Module = fx.Module(
	"registration",
	fx.Provide(
		fx.Annotate(
			New,
			fx.As(new(port.RegistrationStore)),
		),
	),
)
