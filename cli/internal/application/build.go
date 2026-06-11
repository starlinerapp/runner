package application

import (
	"fmt"

	"starliner.app/runner/internal/domain/port"
)

type BuildApplication struct {
	git      port.Git
	buildkit port.Buildkit
}

func NewBuildApplication(
	git port.Git,
	buildkit port.Buildkit,
) *BuildApplication {
	return &BuildApplication{
		git:      git,
		buildkit: buildkit,
	}
}

func (ba *BuildApplication) BuildDockerImage(
	repository string,
	githubToken string,
	dockerfile string,
	buildContext string,
	registryUrl string,
	registryUsername string,
	registryPassword string,
) error {
	workspace, err := ba.git.Checkout(repository, "main", githubToken)
	if err != nil {
		return err
	}
	defer func() {
		_ = workspace.Close()
	}()

	logs, err := ba.buildkit.BuildAndPublish(
		buildContext,
		dockerfile,
		registryUrl,
		registryUsername,
		registryPassword,
		workspace.CommitSHA(),
		nil,
	)
	fmt.Printf("%s", logs)

	return nil
}
