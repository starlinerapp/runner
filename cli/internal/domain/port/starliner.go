package port

type Starliner interface {
	RegisterRunner(token string) error
}
