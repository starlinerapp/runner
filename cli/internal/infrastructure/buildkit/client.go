package buildkit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/tonistiigi/fsutil"
	"starliner.app/runner/internal/domain/port"
)

type Client struct {
	vm port.VM
}

func NewClient(vm port.VM) *Client {
	return &Client{
		vm: vm,
	}
}

func (c *Client) BuildAndPublish(
	projectDir string,
	dockerfilePath string,
	registryUrl string,
	registryUsername string,
	registryPassword string,
	imageTag string,
	args []*port.Arg,
) (string, error) {
	ctx := context.Background()

	vm, err := c.vm.CreateVM()
	if err != nil {
		return "", fmt.Errorf("failed to create VM: %w", err)
	}
	defer func() {
		_ = c.vm.DeleteVM(vm.ID)
	}()

	addr := fmt.Sprintf("tcp://%s:1234", vm.GuestIP)
	bkClient, err := client.New(ctx, addr)

	if err != nil {
		return "", fmt.Errorf("failed to connect to buildkit: %w", err)
	}
	defer func(bkClient *client.Client) {
		if err := bkClient.Close(); err != nil {
			log.Printf("failed to close buildkit client: %v", err)
		}
	}(bkClient)

	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project dir: %w", err)
	}

	dockerfileRelPath := dockerfilePath
	if filepath.IsAbs(dockerfilePath) {
		dockerfileRelPath, err = filepath.Rel(absProjectDir, dockerfilePath)
		if err != nil {
			return "", fmt.Errorf("failed to compute relative dockerfile path: %w", err)
		}
	}

	dockerfileRelPath = filepath.Clean(dockerfileRelPath)

	if dockerfileRelPath == "." {
		return "", fmt.Errorf("dockerfile path points to the context directory, not a file")
	}

	if filepath.IsAbs(dockerfileRelPath) ||
		dockerfileRelPath == ".." ||
		strings.HasPrefix(dockerfileRelPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf(
			"dockerfile must be inside build context: projectDir=%s dockerfilePath=%s resolvedDockerfile=%s",
			absProjectDir,
			dockerfilePath,
			dockerfileRelPath,
		)
	}

	contextFS, err := fsutil.NewFS(absProjectDir)
	if err != nil {
		return "", fmt.Errorf("failed to create context FS: %w", err)
	}

	frontendAttrs := map[string]string{
		"filename": filepath.ToSlash(dockerfileRelPath),
		"platform": "linux/amd64",
	}

	for _, a := range args {
		if a == nil {
			continue
		}
		frontendAttrs["build-arg:"+a.Name] = a.Value
	}

	dockerConfig := &configfile.ConfigFile{
		AuthConfigs: map[string]types.AuthConfig{
			registryUrl: {
				Username: registryUsername,
				Password: registryPassword,
			},
		},
	}

	attachable := []session.Attachable{
		authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
			AuthConfigProvider: authprovider.LoadAuthConfig(dockerConfig),
		}),
	}

	statusCh := make(chan *client.SolveStatus)
	doneCh := make(chan struct{})

	var (
		logs strings.Builder
		mu   sync.Mutex
	)

	_, buildErr := bkClient.Solve(
		ctx,
		nil,
		client.SolveOpt{
			Frontend:      "dockerfile.v0",
			FrontendAttrs: frontendAttrs,
			Exports: []client.ExportEntry{
				{
					Type: client.ExporterImage,
					Attrs: map[string]string{
						"name":              imageTag,
						"push":              "true",
						"oci-mediatypes":    "false",
						"compression":       "gzip",
						"force-compression": "true",
					},
				},
			},
			LocalMounts: map[string]fsutil.FS{
				"context":    contextFS,
				"dockerfile": contextFS,
			},
			Session: attachable,
		},
		statusCh,
	)

	<-doneCh

	mu.Lock()
	logOutput := logs.String()
	mu.Unlock()

	if buildErr != nil {
		if errors.Is(buildErr, context.Canceled) {
			return logOutput, buildErr
		}
		return logOutput, fmt.Errorf("build and push failed: %w", buildErr)
	}
	return logOutput, nil
}
