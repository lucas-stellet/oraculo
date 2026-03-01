// src/output/json.go
package output

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/lucas/oraculo/src/domain"
)

// WriteJSON marshals v to indented JSON and writes it to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteError maps a domain error to the documented JSON error format and writes to w.
func WriteError(w io.Writer, err error) {
	code := "unknown_error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		code = "not_found"
	case errors.Is(err, domain.ErrAlreadyExists):
		code = "already_exists"
	case errors.Is(err, domain.ErrInvalidTransition):
		code = "invalid_transition"
	case errors.Is(err, domain.ErrCyclicDependency):
		code = "cyclic_dependency"
	case errors.Is(err, domain.ErrMissingRequired):
		code = "missing_required"
	case errors.Is(err, domain.ErrInvalidPhase):
		code = "invalid_phase"
	}
	WriteJSON(w, map[string]string{
		"error":   code,
		"message": err.Error(),
	})
}
