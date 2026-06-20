package port

type RunnerRegistration struct {
	Token             string
	MaxConcurrentJobs int
}

type RegistrationStore interface {
	SaveRegistration(registration RunnerRegistration) error
	Token() (string, error)
	MaxConcurrentJobs() (int, error)
}
