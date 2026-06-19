package port

type Starliner interface {
	RegisterRunner(token, name string, labels []string, maxConcurrentJobs int, insecureSkipTLSVerify bool) error
}
