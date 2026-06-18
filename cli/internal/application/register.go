package application

import "starliner.app/runner/internal/domain/port"

type RegisterApplication struct {
	starliner port.Starliner
}

func NewRegisterApplication(starliner port.Starliner) *RegisterApplication {
	return &RegisterApplication{
		starliner: starliner,
	}
}

func (a *RegisterApplication) RegisterRunner(token string) error {
	return a.starliner.RegisterRunner(token)
}
