package application

import (
	"path/filepath"

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
	branchName string,
	githubToken string,
	dockerfile string,
	buildContext string,
	registryUrl string,
	registryUsername string,
	registryPassword string,
) error {
	workspace, err := ba.git.Checkout(repository, branchName, githubToken)
	if err != nil {
		return err
	}
	defer func() {
		_ = workspace.Close()
	}()

	_, err = ba.buildkit.BuildAndPublish(
		filepath.Join(workspace.Path(), buildContext),
		dockerfile,
		registryUrl,
		registryUsername,
		registryPassword,
		workspace.CommitSHA(),
		nil,
	)

	return err
}
