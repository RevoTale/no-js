package framework

import "errors"

type NotFoundError interface {
	error
	NotFound() bool
}

type defaultNotFoundError struct{}

func (defaultNotFoundError) Error() string {
	return "not found"
}

func (defaultNotFoundError) NotFound() bool {
	return true
}

var ErrNotFound error = defaultNotFoundError{}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var target NotFoundError
	return errors.As(err, &target) && target.NotFound()
}
