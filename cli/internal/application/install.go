package application

import "starliner.app/runner/internal/infrastructure/bundle"

type InstallApplication struct{}

func NewInstallApplication() *InstallApplication {
	return &InstallApplication{}
}

func (a *InstallApplication) Install() error {
	return bundle.Install()
}
