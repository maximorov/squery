package squery

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// Select creates a new squirrel.SelectBuilder with the given columns and sets the placeholder format to @p.
func Select(columns ...string) sq.SelectBuilder {
	return sq.Select(columns...).PlaceholderFormat(sq.AtP)
}

// SelectAsStruct creates a new squirrel.SelectBuilder that selects a STRUCT from the given columns.
func SelectAsStruct(columnName string, columns ...string) sq.SelectBuilder {
	return Select(fmt.Sprintf("AS STRUCT %s", strings.Join(columns, `,`))).
		Prefix(`ARRAY(`).
		Suffix(fmt.Sprintf(`) as %s`, columnName))
}

// SelectColumn creates a new squirrel.SelectBuilder that selects a single column.
func SelectColumn(columnName string, column string) sq.SelectBuilder {
	return Select(column).
		Prefix(`(`).
		Suffix(fmt.Sprintf(`) as %s`, columnName))
}
