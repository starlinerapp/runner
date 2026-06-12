package application

import (
	"path/filepath"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
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
	registryUsername string,
	registryPassword string,
	image string,
) error {
	workspace, err := ba.git.Checkout(repository, branchName, githubToken)
	if err != nil {
		return err
	}
	defer func() {
		_ = workspace.Close()
	}()

	imageRef, err := value.ParseImageRef(image)
	if err != nil {
		return err
	}
	imageRef = imageRef.WithTag(workspace.CommitSHA())

	_, err = ba.buildkit.BuildAndPublish(
		filepath.Join(workspace.Path(), buildContext),
		dockerfile,
		registryUsername,
		registryPassword,
		imageRef,
		nil,
	)

	return err
}
