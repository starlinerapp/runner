package application

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"starliner.app/runner/internal/domain/port"
	"starliner.app/runner/internal/domain/value"
)

const claimJobInterval = 2 * time.Second

type WorkerApplication struct {
	registration port.RegistrationStore
	jobs         port.JobGateway
	heartbeat    port.HeartbeatGateway
	workload     port.WorkloadTracker
	executor     port.JobExecutor
	claimMu      sync.Mutex
	inflight     sync.Map // build ID -> struct{}
}

func NewWorkerApplication(
	registration port.RegistrationStore,
	jobs port.JobGateway,
	heartbeat port.HeartbeatGateway,
	workload port.WorkloadTracker,
	executor port.JobExecutor,
) *WorkerApplication {
	return &WorkerApplication{
		registration: registration,
		jobs:         jobs,
		heartbeat:    heartbeat,
		workload:     workload,
		executor:     executor,
	}
}

func (a *WorkerApplication) Run(ctx context.Context, insecureSkipTLSVerify bool) error {
	token, err := a.registration.Token()
	if err != nil {
		return err
	}
	if token == "" {
		return fmt.Errorf("runner not registered; run `runner register` first")
	}

	jobSession, err := a.jobs.OpenSession(ctx, insecureSkipTLSVerify)
	if err != nil {
		return err
	}
	defer func() { _ = jobSession.Close() }()

	heartbeatSession, err := a.heartbeat.OpenHeartbeatSession(ctx, insecureSkipTLSVerify)
	if err != nil {
		return err
	}
	defer func() { _ = heartbeatSession.Close() }()

	errCh := make(chan error, 2)
	go func() {
		errCh <- heartbeatSession.Run(ctx)
	}()
	go func() {
		errCh <- a.pollJobs(ctx, jobSession)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *WorkerApplication) pollJobs(ctx context.Context, session port.JobSession) error {
	ticker := time.NewTicker(claimJobInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = a.claimAndRunJob(ctx, session)
		}
	}
}

func (a *WorkerApplication) claimAndRunJob(ctx context.Context, session port.JobSession) error {
	a.claimMu.Lock()
	defer a.claimMu.Unlock()

	if a.workload != nil {
		if a.maxCapacityReached() {
			return nil
		}
		a.workload.Increment()
	}

	job, err := session.ClaimJob(ctx)
	if err != nil {
		if a.workload != nil {
			a.workload.Decrement()
		}
		log.Printf("claim job: %v", err)
		return nil
	}
	if job == nil {
		if a.workload != nil {
			a.workload.Decrement()
		}
		return nil
	}

	if _, loaded := a.inflight.LoadOrStore(job.BuildID, struct{}{}); loaded {
		if a.workload != nil {
			a.workload.Decrement()
		}
		return nil
	}

	go func(j value.BuildJob) {
		defer a.inflight.Delete(j.BuildID)
		a.executeJob(ctx, session, j)
	}(*job)

	return nil
}

func (a *WorkerApplication) maxCapacityReached() bool {
	maxJobs, err := a.registration.MaxConcurrentJobs()
	if err != nil {
		return false
	}

	return a.workload.ActiveJobs() >= maxJobs
}

func (a *WorkerApplication) executeJob(
	ctx context.Context,
	session port.JobSession,
	job value.BuildJob,
) {
	if a.workload != nil {
		defer a.workload.Decrement()
	}

	reporter := &jobReporter{
		session: session,
		job:     job,
	}

	runnerName, _ := a.registration.Name()
	if runnerName != "" {
		reporter.PublishLog(fmt.Sprintf("Picked up by runner %s\n", runnerName))
	}

	if a.executor == nil {
		msg := "job executor not configured\n"
		reporter.PublishLog(msg)
		reporter.SendResult(value.BuildResult{
			BuildID:      job.BuildID,
			DeploymentID: job.DeploymentID,
			Logs:         msg,
			Status:       value.BuildStatusFailed,
		})
		return
	}

	if err := a.executor.ExecuteJob(ctx, job, reporter); err != nil {
		msg := fmt.Sprintf("build failed: %v\n", err)
		reporter.PublishLog(msg)
		reporter.SendResult(value.BuildResult{
			BuildID:      job.BuildID,
			DeploymentID: job.DeploymentID,
			Logs:         msg,
			Status:       value.BuildStatusFailed,
		})
		return
	}
}

type jobReporter struct {
	session port.JobSession
	job     value.BuildJob
}

func (r *jobReporter) PublishLog(line string) {
	_ = r.session.ReportBuildLog(context.Background(), r.job.BuildID, []byte(line))
}

func (r *jobReporter) SendResult(result value.BuildResult) {
	_ = r.session.ReportBuildResult(context.Background(), result)
}
