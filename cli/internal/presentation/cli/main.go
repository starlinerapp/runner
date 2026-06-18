package cli

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"starliner.app/runner/internal/presentation/cli/handler"
)

func Register(
	lc fx.Lifecycle,
	sd fx.Shutdowner,
	register *handler.RegisterHandler,
	build *handler.BuildHandler,
	install *handler.InstallHandler,
	vm *handler.VMHandler,
) {
	rootCmd := &cobra.Command{
		Version:       "0.0.1",
		Use:           "runner",
		Example:       "runner",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		register.NewRegisterCmd(),
		build.NewBuildCmd(),
		install.NewInstallCmd(),
		vm.NewVMCmd(),
	)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				err := rootCmd.ExecuteContext(context.Background())
				_ = sd.Shutdown(fx.ExitCode(exitCodeFor(err)))
			}()
			return nil
		},
	})
}

func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	return 1
}
