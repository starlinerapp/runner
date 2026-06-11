package git

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"starliner.app/runner/internal/domain/port"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Checkout(repoUrl string, branchName string, accessToken string) (port.Workspace, error) {
	path, err := os.MkdirTemp("", "runner-workspace-*")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Cloning %s (branch %s)...\n", repoUrl, branchName)
	commitSHA, err := clone(path, repoUrl, branchName, accessToken)
	if err != nil {
		_ = os.RemoveAll(path)
		return nil, err
	}
	fmt.Printf("Cloned at %s\n", commitSHA)

	return NewWorkspace(path, commitSHA), nil
}

func clone(dir string, repoUrl string, branchName string, accessToken string) (commitSHA string, err error) {
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: repoUrl,
		Auth: &http.BasicAuth{
			Username: "x-access-token",
			Password: accessToken,
		},
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return "", err
	}
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), err
}
