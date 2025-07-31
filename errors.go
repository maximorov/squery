package squery

import (
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/go-errors/errors"
)

func ParentIsMissing(err error) bool {
	return strings.Contains(err.Error(), `is missing. Row cannot`)
}

func NilIfNotFound(err error) error {
	if errors.Is(err, spanner.ErrRowNotFound) {
		return nil
	}
	return err
}
