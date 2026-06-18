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

func (a *RegisterApplication) RegisterRunner(token string, insecureSkipTLSVerify bool) error {
	if err := a.starliner.RegisterRunner(token, insecureSkipTLSVerify); err != nil {
		return err
	}

	return a.credentials.SaveToken(token)
}
