package services

type NotFoundError struct {
	s string
}

func (e NotFoundError) Error() string {
	return e.s
}
