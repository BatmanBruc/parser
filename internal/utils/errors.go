package utils

import "fmt"

type AppError struct {
	Service string
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %v", e.Service, e.Err)
}

func NewError(service string, err error) *AppError {
	return &AppError{
		Service: service,
		Err:     err,
	}
}
