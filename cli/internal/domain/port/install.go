package port

type Installer interface {
	Install(baseURL string) error
}
