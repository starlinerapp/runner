package port

type Arg struct {
	Name  string
	Value string
}

type Buildkit interface {
	BuildAndPublish(
		projectDir string,
		dockerfilePath string,
		registryUrl string,
		registryUsername string,
		registryPassword string,
		imageTag string,
		args []*Arg,
	) (string, error)
}
