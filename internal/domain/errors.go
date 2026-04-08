package domain

import "errors"

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrFatal indicates an error that should not be retried.
var ErrFatal = errors.New("fatal error")
