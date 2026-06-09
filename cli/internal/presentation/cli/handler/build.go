package handler

import "github.com/spf13/cobra"

type BuildHandler struct{}

func NewBuildHandler() *BuildHandler {
	return &BuildHandler{}
}

type BuildOpts struct {
	Repository string
	Dockerfile string
	Context    string
	Registry   string
}

func (bh *BuildHandler) NewBuildCmd() *cobra.Command {
	opts := &BuildOpts{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a Docker image",
		Long:  "Build a Docker image from a Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Repository, "repository", "", "Git repository URL")
	cmd.Flags().StringVar(&opts.Dockerfile, "dockerfile", "Dockerfile", "Path to Dockerfile relative to context or repository root")
	cmd.Flags().StringVar(&opts.Context, "context", ".", "Build context directory relative to repository root")
	cmd.Flags().StringVar(&opts.Registry, "registry", "", "Docker registry to push the image to")

	_ = cmd.MarkFlagRequired("repository")
	_ = cmd.MarkFlagRequired("registry")

	return cmd
}
