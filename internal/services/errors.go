package services

import "errors"

type NotFoundError struct {
	s string
}

func (e NotFoundError) Error() string {
	return e.s
}

var ErrInvalidDateFormat = errors.New("invalid date format, expected YYYY-MM-DD")
