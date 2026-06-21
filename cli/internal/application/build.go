package application

import (
	"context"
	"fmt"
	"path"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
)

type BuildApplication struct {
	git      port.Git
	vm       port.VM
	buildkit port.Buildkit
}

func NewBuildApplication(
	git port.Git,
	vm port.VM,
	buildkit port.Buildkit,
) *BuildApplication {
	return &BuildApplication{
		git:      git,
		vm:       vm,
		buildkit: buildkit,
	}
}

func (a *BuildApplication) ExecuteJob(
	ctx context.Context,
	job value.BuildJob,
	reporter port.JobReporter,
) error {
	if reporter == nil {
		reporter = &noopJobReporter{}
	}

	reporter.PublishLog(fmt.Sprintf("Cloning %s (branch %s)...\n", job.GitURL, job.BranchName))

	workspace, err := a.git.Checkout(job.GitURL, job.BranchName, job.AccessToken)
	if err != nil {
		return err
	}
	defer func() {
		_ = workspace.Close()
	}()

	commitHash := workspace.CommitSHA()

	imageRef, err := value.ParseImageRef(path.Join(job.ImageRegistry, job.ImageName))
	if err != nil {
		return err
	}
	tag := imageRef.WithTag(commitHash).String()

	buildContext := job.RootDirectory
	if buildContext == "" {
		buildContext = "."
	}

	guest, err := a.vm.CreateVM()
	if err != nil {
		return fmt.Errorf("create VM: %w", err)
	}
	defer func() {
		reporter.PublishLog(fmt.Sprintf("tearing down microVM %s\n", guest.ID))
		_ = a.vm.DeleteVM(guest.ID)
	}()

	args := make([]*port.Arg, 0, len(job.Args))
	for _, arg := range job.Args {
		args = append(args, &port.Arg{
			Name:  arg.Name,
			Value: arg.Value,
		})
	}

	logs, err := a.buildkit.BuildAndPublish(
		*guest,
		path.Join(workspace.Path(), buildContext),
		job.DockerfilePath,
		job.RegistryToken,
		imageRef.WithTag(commitHash),
		args,
		reporter.PublishLog,
	)
	if err != nil {
		a.vm.Diagnose(*guest)
		return err
	}

	reporter.SendResult(value.BuildResult{
		BuildID:      job.BuildID,
		DeploymentID: job.DeploymentID,
		CommitHash:   commitHash,
		Tag:          tag,
		ImageName:    imageRef.String(),
		Logs:         logs,
		Status:       value.BuildStatusSuccess,
	})

	return nil
}

type noopJobReporter struct{}

func (noopJobReporter) PublishLog(string)            {}
func (noopJobReporter) SendResult(value.BuildResult) {}
