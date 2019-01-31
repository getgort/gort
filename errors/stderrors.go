package errors

import (
	"errors"
)

var (
	// ErrIO is returned when underlying IO error is received.
	ErrIO = errors.New("IO error")

	// ErrMarshal is returned when an item can't be marshaled into JSON/YAML.
	ErrMarshal = errors.New("marshalling error")

	// ErrUnmarshal is returned when an item can't be unmarshaled from JSON/YAML.
	ErrUnmarshal = errors.New("unmarshalling error")
)
