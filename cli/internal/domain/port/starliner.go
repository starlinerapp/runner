package port

import "context"

type Starliner interface {
	RegisterRunner(token, name string, labels []string, maxConcurrentJobs int, insecureSkipTLSVerify bool) error
	ServeHeartbeats(ctx context.Context, insecureSkipTLSVerify bool) error
}
