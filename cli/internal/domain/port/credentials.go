package port

type CredentialsStore interface {
	SaveToken(token string) error
	Token() (string, error)
}
