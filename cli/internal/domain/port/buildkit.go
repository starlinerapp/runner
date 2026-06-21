package port

import "starliner.app/runner/internal/domain/value"

type Arg struct {
	Name  string
	Value string
}

type Buildkit interface {
	BuildAndPublish(
		guest value.VM,
		projectDir string,
		dockerfilePath string,
		registryToken string,
		imageRef value.ImageRef,
		args []*Arg,
		publishLog LogPublisher,
	) (string, error)
}
