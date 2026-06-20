package handler

import (
	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
)

type RunHandler struct {
	runApplication *application.RunApplication
}

func NewRunHandler(runApplication *application.RunApplication) *RunHandler {
	return &RunHandler{
		runApplication: runApplication,
	}
}

type RunOpts struct {
	InsecureSkipTLSVerify bool
}

func (h *RunHandler) NewRunCmd() *cobra.Command {
	opts := &RunOpts{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the runner daemon",
		Long:  "Star the runner daemon",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return h.runApplication.Run(opts.InsecureSkipTLSVerify)
		},
	}

	cmd.Flags().BoolVar(
		&opts.InsecureSkipTLSVerify,
		"insecure-skip-tls-verify",
		false,
		"skip TLS certificate verification",
	)

	return cmd
}
