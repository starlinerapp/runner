package buildkit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"
	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
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
	guest value.VM,
	projectDir string,
	dockerfilePath string,
	registryToken string,
	imageRef value.ImageRef,
	args []*port.Arg,
	publishLog port.LogPublisher,
) (string, error) {
	ctx := context.Background()

	waitMsg := fmt.Sprintf("Waiting for buildkit at %s:%d...\n", guest.GuestIP, Port)
	publishLog(waitMsg)
	if err := Wait(guest.GuestIP, ConnectTimeout); err != nil {
		c.vm.Diagnose(guest)
		return "", err
	}

	addr := fmt.Sprintf("tcp://%s:%d", guest.GuestIP, Port)
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

	if registryToken == "" {
		return "", fmt.Errorf("registry push token is required")
	}

	authConfigs := map[string]types.AuthConfig{}
	if host := imageRef.RegistryHost(); host != "" {
		authConfigs[host] = types.AuthConfig{
			RegistryToken: registryToken,
		}
	}

	dockerConfig := &configfile.ConfigFile{
		AuthConfigs: authConfigs,
	}

	attachable := []session.Attachable{
		authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
			AuthConfigProvider: authprovider.LoadAuthConfig(dockerConfig),
		}),
	}

	cacheRef := registryCacheRef(imageRef)

	statusCh := make(chan *client.SolveStatus)
	solveOpt := client.SolveOpt{
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterImage,
				Attrs: map[string]string{
					"name":              imageRef.String(),
					"push":              "true",
					"oci-mediatypes":    "false",
					"compression":       "gzip",
					"force-compression": "true",
				},
			},
		},
		CacheExports: []client.CacheOptionsEntry{
			{
				Type: "registry",
				Attrs: map[string]string{
					"ref":  cacheRef,
					"mode": "max",
				},
			},
		},
		CacheImports: []client.CacheOptionsEntry{
			{
				Type: "registry",
				Attrs: map[string]string{
					"ref": cacheRef,
				},
			},
		},
		LocalMounts: map[string]fsutil.FS{
			"context":    contextFS,
			"dockerfile": contextFS,
		},
		Session: attachable,
	}

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := bkClient.Solve(egCtx, nil, solveOpt, statusCh)
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("build and push failed: %w", err)
	})

	logOut := &logWriter{publish: publishLog}

	eg.Go(func() error {
		d, err := progressui.NewDisplay(logOut, progressui.PlainMode)
		if err != nil {
			return err
		}
		_, err = d.UpdateFrom(context.TODO(), statusCh)
		return err
	})

	if err := eg.Wait(); err != nil {
		return logOut.String(), err
	}
	return logOut.String(), nil
}

type logWriter struct {
	buf     strings.Builder
	publish port.LogPublisher
}

func (w *logWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	if err != nil {
		return n, err
	}
	if len(p) > 0 {
		w.publish(string(p))
	}
	return len(p), nil
}

func (w *logWriter) String() string {
	return w.buf.String()
}

const registryBuildCacheTag = "buildcache"

func registryCacheRef(imageRef value.ImageRef) string {
	name := imageRef.String()
	if i := strings.LastIndex(name, ":"); i != -1 && strings.LastIndex(name, "/") < i {
		return name[:i] + ":" + registryBuildCacheTag
	}
	return name + ":" + registryBuildCacheTag
}
