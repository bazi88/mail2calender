package common

import "errors"

var (
	// ErrNotFound indicates that a requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates that the input data is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates that the user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInternalServer indicates an internal server error
	ErrInternalServer = errors.New("internal server error")
)
