package application

import "starliner.app/runner/internal/domain/port"

type InstallApplication struct {
	installer port.Installer
}

func NewInstallApplication(installer port.Installer) *InstallApplication {
	return &InstallApplication{installer: installer}
}

func (a *InstallApplication) Install(baseURL string) error {
	return a.installer.Install(baseURL)
}
