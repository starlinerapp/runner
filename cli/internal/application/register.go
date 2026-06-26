package application

import "starliner.app/runner/internal/domain/port"

type RegisterApplication struct {
	starliner    port.Starliner
	registration port.RegistrationStore
}

func NewRegisterApplication(
	starliner port.Starliner,
	registration port.RegistrationStore,
) *RegisterApplication {
	return &RegisterApplication{
		starliner:    starliner,
		registration: registration,
	}
}

type RegisterRunnerInput struct {
	Token             string
	Name              string
	Labels            []string
	MaxConcurrentJobs int
}

func (a *RegisterApplication) RegisterRunner(input RegisterRunnerInput, insecureSkipTLSVerify bool) error {
	if err := a.starliner.RegisterRunner(
		input.Token,
		input.Name,
		input.Labels,
		input.MaxConcurrentJobs,
		insecureSkipTLSVerify,
	); err != nil {
		return err
	}

	return a.registration.SaveRegistration(port.RunnerRegistration{
		Token:             input.Token,
		Name:              input.Name,
		MaxConcurrentJobs: input.MaxConcurrentJobs,
	})
}
