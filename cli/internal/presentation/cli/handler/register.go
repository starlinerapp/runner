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
	Token                 string
	InsecureSkipTLSVerify bool
}

func (h *RegisterHandler) NewRegisterCmd() *cobra.Command {
	opts := &RegisterOpts{}

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register runner with Starliner",
		Long:  "Register runner with Starliner Control Plane.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return h.registerApplication.RegisterRunner(opts.Token, opts.InsecureSkipTLSVerify)
		},
	}

	cmd.Flags().StringVar(&opts.Token, "token", "", "token")
	cmd.Flags().BoolVar(&opts.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "skip TLS certificate verification")

	_ = cmd.MarkFlagRequired("token")

	return cmd
}
