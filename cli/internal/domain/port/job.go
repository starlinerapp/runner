package port

import (
	"context"

	"starliner.app/runner/internal/domain/value"
)

type LogPublisher func(line string)

type JobReporter interface {
	PublishLog(line string)
	SendResult(result value.BuildResult)
}

type JobExecutor interface {
	ExecuteJob(ctx context.Context, job value.BuildJob, reporter JobReporter) error
}

type JobSession interface {
	ClaimJob(ctx context.Context) (*value.BuildJob, error)
	ReportBuildLog(ctx context.Context, buildID int64, data []byte) error
	ReportBuildResult(ctx context.Context, result value.BuildResult) error
	Close() error
}

type JobGateway interface {
	OpenSession(ctx context.Context, insecureSkipTLSVerify bool) (JobSession, error)
}
