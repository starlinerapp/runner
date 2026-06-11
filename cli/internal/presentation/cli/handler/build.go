package handler

import (
	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
)

type BuildHandler struct {
	buildApplication *application.BuildApplication
}

func NewBuildHandler(
	buildApplication *application.BuildApplication,
) *BuildHandler {
	return &BuildHandler{
		buildApplication: buildApplication,
	}
}

type BuildOpts struct {
	Repository       string
	Dockerfile       string
	Context          string
	Registry         string
	RegistryUsername string
	RegistryPassword string
	GithubToken      string
}

func (bh *BuildHandler) NewBuildCmd() *cobra.Command {
	opts := &BuildOpts{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a Docker image",
		Long:  "Build a Docker image from a Git repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bh.buildApplication.BuildDockerImage(
				opts.Repository,
				opts.GithubToken,
				opts.Dockerfile,
				opts.Context,
				opts.Registry,
				opts.RegistryUsername,
				opts.RegistryPassword,
			)
		},
	}

	cmd.Flags().StringVar(&opts.Repository, "repository", "", "Git repository URL")
	cmd.Flags().StringVar(&opts.GithubToken, "github-token", "", "GitHub access token")
	cmd.Flags().StringVar(&opts.Dockerfile, "dockerfile", "Dockerfile", "Path to Dockerfile relative to context or repository root")
	cmd.Flags().StringVar(&opts.Context, "context", ".", "Build context directory relative to repository root")
	cmd.Flags().StringVar(&opts.Registry, "registry", "", "Docker registry to push the image to")
	cmd.Flags().StringVar(&opts.RegistryUsername, "registry-username", "", "Docker registry username")
	cmd.Flags().StringVar(&opts.RegistryPassword, "registry-password", "", "Docker registry password")

	_ = cmd.MarkFlagRequired("repository")
	_ = cmd.MarkFlagRequired("github-token")
	_ = cmd.MarkFlagRequired("registry")

	return cmd
}
