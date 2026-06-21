package port

type RunnerRegistration struct {
	Token             string
	Name              string
	MaxConcurrentJobs int
}

type RegistrationStore interface {
	SaveRegistration(registration RunnerRegistration) error
	Token() (string, error)
	Name() (string, error)
	MaxConcurrentJobs() (int, error)
}
