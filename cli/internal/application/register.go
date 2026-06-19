package application

import "starliner.app/runner/internal/domain/port"

type RegisterApplication struct {
	starliner   port.Starliner
	credentials port.CredentialsStore
}

func NewRegisterApplication(
	starliner port.Starliner,
	credentials port.CredentialsStore,
) *RegisterApplication {
	return &RegisterApplication{
		starliner:   starliner,
		credentials: credentials,
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

	return a.credentials.SaveToken(input.Token)
}
