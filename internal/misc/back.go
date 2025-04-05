package misc

type ErrorBack[T any] struct {
	Content   T
	ErrorChan chan error
}
