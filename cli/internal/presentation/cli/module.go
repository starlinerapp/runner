package cli

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/presentation/cli/handler"
)

var Module = fx.Module(
	"cli",
	handler.Module,
	fx.Invoke(Register),
)
