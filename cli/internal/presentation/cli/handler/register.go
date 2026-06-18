package handler

import (
	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
)

type RegisterHandler struct {
	registerApplication *application.RegisterApplication
}

func NewRegisterHandler(registerApplication *application.RegisterApplication) *RegisterHandler {
	return &RegisterHandler{
		registerApplication: registerApplication,
	}
}

type RegisterOpts struct {
	Token string
}

func (h *RegisterHandler) NewRegisterCmd() *cobra.Command {
	opts := &RegisterOpts{}

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register runner with Starliner",
		Long:  "Register runner with Starliner Control Plane.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return h.registerApplication.RegisterRunner(opts.Token)
		},
	}

	cmd.Flags().StringVar(&opts.Token, "token", "", "token")

	_ = cmd.MarkFlagRequired("token")

	return cmd
}
