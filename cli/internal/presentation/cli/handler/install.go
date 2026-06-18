package handler

import (
	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
)

type InstallHandler struct {
	installApplication *application.InstallApplication
}

type InstallOpts struct {
	BaseURL string
}

func NewInstallHandler(
	installApplication *application.InstallApplication,
) *InstallHandler {
	return &InstallHandler{
		installApplication: installApplication,
	}
}

func (h *InstallHandler) NewInstallCmd() *cobra.Command {
	opts := &InstallOpts{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install runner to /usr/local/bin",
		Long:  "Install the runner binary and VM assets.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return h.installApplication.Install(opts.BaseURL)
		},
	}

	cmd.Flags().StringVar(&opts.BaseURL, "base-url", "", "Starliner control plane base URL")
	_ = cmd.MarkFlagRequired("base-url")

	return cmd
}
