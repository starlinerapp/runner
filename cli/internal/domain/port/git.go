package port

type Workspace interface {
	Path() string
	CommitSHA() string
	Close() error
}

type Git interface {
	Checkout(repoUrl string, branchName string, accessToken string) (Workspace, error)
}
