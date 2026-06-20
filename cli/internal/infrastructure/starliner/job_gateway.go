package starliner

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
	v1 "starliner.app/runner/internal/infrastructure/grpc/proto/v1"
)

const runnerTokenMetadataKey = "authorization"

type JobGateway struct {
	config       port.ConfigStore
	registration port.RegistrationStore
	workload     port.WorkloadTracker
}

func NewJobGateway(
	config port.ConfigStore,
	registration port.RegistrationStore,
	workload port.WorkloadTracker,
) port.JobGateway {
	return &JobGateway{
		config:       config,
		registration: registration,
		workload:     workload,
	}
}

func (g *JobGateway) OpenSession(ctx context.Context, insecureSkipTLSVerify bool) (port.JobSession, error) {
	token, err := g.registration.Token()
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("runner not registered; run `runner register` first")
	}

	baseURL, err := g.config.BaseURL()
	if err != nil {
		return nil, err
	}

	target, err := grpcTarget(baseURL)
	if err != nil {
		return nil, err
	}

	conn, err := dialRunnerGRPC(target, baseURL, insecureSkipTLSVerify)
	if err != nil {
		return nil, err
	}

	jobClient := v1.NewRunnerJobServiceClient(conn)
	return &jobSession{
		conn:   conn,
		client: jobClient,
		token:  token,
	}, nil
}

type jobSession struct {
	conn   *grpc.ClientConn
	client v1.RunnerJobServiceClient
	token  string
}

func (s *jobSession) ClaimJob(ctx context.Context) (*value.BuildJob, error) {
	callCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(runnerTokenMetadataKey, s.token))
	resp, err := s.client.ClaimJob(callCtx, &v1.ClaimJobRequest{})
	if err != nil {
		return nil, err
	}

	job := resp.GetJob()
	if job == nil {
		return nil, nil
	}

	return mapBuildJob(job), nil
}

func (s *jobSession) ReportBuildLog(ctx context.Context, buildID int64, data []byte) error {
	callCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(runnerTokenMetadataKey, s.token))
	_, err := s.client.ReportBuildLog(callCtx, &v1.ReportBuildLogRequest{
		BuildId: buildID,
		Data:    data,
	})
	return err
}

func (s *jobSession) ReportBuildResult(ctx context.Context, result value.BuildResult) error {
	callCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(runnerTokenMetadataKey, s.token))
	_, err := s.client.ReportBuildResult(callCtx, &v1.ReportBuildResultRequest{
		Result: mapBuildResult(result),
	})
	return err
}

func (s *jobSession) Close() error {
	return s.conn.Close()
}

func mapBuildJob(job *v1.BuildJob) *value.BuildJob {
	args := make([]value.BuildArg, 0, len(job.GetArgs()))
	for _, arg := range job.GetArgs() {
		if arg == nil {
			continue
		}
		args = append(args, value.BuildArg{
			Name:  arg.GetName(),
			Value: arg.GetValue(),
		})
	}

	return &value.BuildJob{
		BuildID:        job.GetBuildId(),
		DeploymentID:   job.GetDeploymentId(),
		ImageName:      job.GetImageName(),
		ImageRegistry:  job.GetImageRegistryUrl(),
		GitURL:         job.GetGitUrl(),
		BranchName:     job.GetBranchName(),
		AccessToken:    job.GetAccessToken(),
		RegistryToken:  job.GetRegistryPushToken(),
		RootDirectory:  job.GetRootDirectory(),
		DockerfilePath: job.GetDockerfilePath(),
		Args:           args,
	}
}

func mapBuildResult(result value.BuildResult) *v1.BuildResult {
	status := v1.BuildStatus_BUILD_STATUS_UNSPECIFIED
	switch result.Status {
	case value.BuildStatusSuccess:
		status = v1.BuildStatus_BUILD_STATUS_SUCCESS
	case value.BuildStatusFailed:
		status = v1.BuildStatus_BUILD_STATUS_FAILED
	default:
		panic("unhandled default case")
	}

	return &v1.BuildResult{
		BuildId:      result.BuildID,
		DeploymentId: result.DeploymentID,
		CommitHash:   result.CommitHash,
		Tag:          result.Tag,
		ImageName:    result.ImageName,
		Logs:         result.Logs,
		Status:       status,
	}
}
