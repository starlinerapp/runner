package git

import (
	"os"
)

type Workspace struct {
	path      string
	commitSHA string
}

func NewWorkspace(path string, commitSHA string) *Workspace {
	return &Workspace{
		path:      path,
		commitSHA: commitSHA,
	}
}

func (w *Workspace) Path() string {
	return w.path
}

func (w *Workspace) CommitSHA() string {
	return w.commitSHA
}

func (w *Workspace) Close() error {
	if w.path == "" {
		return nil
	}

	return os.RemoveAll(w.path)
}
