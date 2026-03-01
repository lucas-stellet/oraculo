// src/domain/errors.go
package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrCyclicDependency  = errors.New("cyclic dependency in task graph")
	ErrMissingRequired   = errors.New("missing required field")
)
