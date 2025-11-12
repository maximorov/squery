package squery

import (
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/go-errors/errors"
)

// ParentIsMissing returns true if the error is a Spanner error indicating a missing parent row.
func ParentIsMissing(err error) bool {
	return strings.Contains(err.Error(), `is missing. Row cannot`)
}

// NilIfNotFound returns nil if the error is a Spanner "row not found" error, otherwise it returns the original error.
func NilIfNotFound(err error) error {
	if errors.Is(err, spanner.ErrRowNotFound) {
		return nil
	}
	return err
}
