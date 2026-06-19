package handler

import (
	"fmt"

	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
	"starliner.app/runner/internal/presentation/cli/prompt"
)

var defaultRegisterLabels = []string{"self-hosted"}

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
	Name                  string
	Labels                []string
	MaxConcurrentJobs     int
	InsecureSkipTLSVerify bool
}

func (h *RegisterHandler) NewRegisterCmd() *cobra.Command {
	opts := &RegisterOpts{}

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register runner with Starliner",
		Long:  "Register runner with Starliner Control Plane.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			input, err := h.resolveRegisterInput(cmd, opts)
			if err != nil {
				return err
			}

			return h.registerApplication.RegisterRunner(input, opts.InsecureSkipTLSVerify)
		},
	}

	cmd.Flags().StringVar(&opts.Token, "token", "", "token")
	cmd.Flags().StringVar(&opts.Name, "name", "", "runner name")
	cmd.Flags().StringSliceVar(&opts.Labels, "labels", nil, "runner labels")
	cmd.Flags().IntVar(&opts.MaxConcurrentJobs, "max-concurrent-jobs", 0, "maximum concurrent jobs")
	cmd.Flags().BoolVar(&opts.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "skip TLS certificate verification")

	_ = cmd.MarkFlagRequired("token")

	return cmd
}

func (h *RegisterHandler) resolveRegisterInput(cmd *cobra.Command, opts *RegisterOpts) (application.RegisterRunnerInput, error) {
	in := cmd.InOrStdin()
	out := cmd.OutOrStdout()

	name := opts.Name
	if name == "" {
		var err error
		name, err = prompt.String(in, out, "Name", "")
		if err != nil {
			return application.RegisterRunnerInput{}, err
		}
	}

	labels := opts.Labels
	if !cmd.Flags().Changed("labels") {
		var err error
		labels, err = prompt.Labels(in, out, defaultRegisterLabels)
		if err != nil {
			return application.RegisterRunnerInput{}, err
		}
	}

	maxConcurrentJobs := opts.MaxConcurrentJobs
	if !cmd.Flags().Changed("max-concurrent-jobs") {
		var err error
		maxConcurrentJobs, err = prompt.PositiveInt(in, out, "Max concurrent jobs")
		if err != nil {
			return application.RegisterRunnerInput{}, err
		}
	} else if maxConcurrentJobs < 1 {
		return application.RegisterRunnerInput{}, fmt.Errorf("max concurrent jobs must be a positive integer")
	}

	if name == "" {
		return application.RegisterRunnerInput{}, fmt.Errorf("name is required")
	}

	return application.RegisterRunnerInput{
		Token:             opts.Token,
		Name:              name,
		Labels:            labels,
		MaxConcurrentJobs: maxConcurrentJobs,
	}, nil
}
