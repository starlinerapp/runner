package main

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/application"
	"starliner.app/runner/internal/infrastructure/firecracker"
	"starliner.app/runner/internal/presentation/cli"
)

func main() {
	fx.New(
		fx.NopLogger,
		firecracker.Module,
		application.Module,
		cli.Module,
	).Run()
}
