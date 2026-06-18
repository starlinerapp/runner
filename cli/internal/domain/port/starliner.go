package port

type Starliner interface {
	RegisterRunner(token string, insecureSkipTLSVerify bool) error
}
