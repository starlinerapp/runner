package main

import (
	"fmt"
	"os"

	"go.uber.org/fx"
	"starliner.app/runner/internal/application"
	"starliner.app/runner/internal/conf"
	"starliner.app/runner/internal/infrastructure/buildkit"
	"starliner.app/runner/internal/infrastructure/bundle"
	"starliner.app/runner/internal/infrastructure/credentials"
	"starliner.app/runner/internal/infrastructure/firecracker"
	"starliner.app/runner/internal/infrastructure/git"
	"starliner.app/runner/internal/infrastructure/registry"
	"starliner.app/runner/internal/infrastructure/starliner"
	"starliner.app/runner/internal/presentation/cli"
)

func main() {
	app := fx.New(
		fx.NopLogger,
		registry.Module,
		credentials.Module,
		buildkit.Module,
		starliner.Module,
		git.Module,
		bundle.Module,
		firecracker.Module,
		application.Module,
		cli.Module,
		conf.Module,
	)
	if err := app.Err(); err != nil {
		if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
			return
		}
		os.Exit(1)
	}
	app.Run()
}
