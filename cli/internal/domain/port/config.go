package port

type ConfigStore interface {
	SaveBaseURL(baseURL string) error
	BaseURL() (string, error)
}
