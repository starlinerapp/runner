package value

type BuildArg struct {
	Name  string
	Value string
}

type BuildJob struct {
	BuildID        int64
	DeploymentID   int64
	ImageName      string
	ImageRegistry  string
	GitURL         string
	BranchName     string
	AccessToken    string
	RegistryToken  string
	RootDirectory  string
	DockerfilePath string
	Args           []BuildArg
}

type BuildStatus int

const (
	BuildStatusUnspecified BuildStatus = iota
	BuildStatusSuccess
	BuildStatusFailed
)

type BuildResult struct {
	BuildID      int64
	DeploymentID int64
	CommitHash   string
	Tag          string
	ImageName    string
	Logs         string
	Status       BuildStatus
}
