package handler

import (
	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
)

type InstallHandler struct {
	installApplication *application.InstallApplication
}

func NewInstallHandler(
	installApplication *application.InstallApplication,
) *InstallHandler {
	return &InstallHandler{
		installApplication: installApplication,
	}
}

func (h *InstallHandler) NewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install runner to /usr/local/bin",
		Long:  "Install the runner binary and VM assets.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return h.installApplication.Install()
		},
	}
}
