package port

import "starliner.app/runner/internal/domain/value"

type Arg struct {
	Name  string
	Value string
}

type Buildkit interface {
	BuildAndPublish(
		projectDir string,
		dockerfilePath string,
		registryUsername string,
		registryPassword string,
		imageRef value.ImageRef,
		args []*Arg,
	) (string, error)
}
