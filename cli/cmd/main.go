package main

import (
	"go.uber.org/fx"
	"starliner.app/runner/internal/application"
	"starliner.app/runner/internal/infrastructure/buildkit"
	"starliner.app/runner/internal/infrastructure/bundle"
	"starliner.app/runner/internal/infrastructure/firecracker"
	"starliner.app/runner/internal/infrastructure/git"
	"starliner.app/runner/internal/infrastructure/registry"
	"starliner.app/runner/internal/presentation/cli"
)

func main() {
	fx.New(
		fx.NopLogger,
		registry.Module,
		buildkit.Module,
		git.Module,
		bundle.Module,
		firecracker.Module,
		application.Module,
		cli.Module,
	).Run()
}
