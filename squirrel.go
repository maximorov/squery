package squery

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func Select(columns ...string) sq.SelectBuilder {
	return sq.Select(columns...).PlaceholderFormat(sq.AtP)
}

func SelectAsStruct(columnName string, columns ...string) sq.SelectBuilder {
	return Select(fmt.Sprintf("AS STRUCT %s", strings.Join(columns, `,`))).
		Prefix(`ARRAY(`).
		Suffix(fmt.Sprintf(`) as %s`, columnName))
}

func SelectColumn(columnName string, column string) sq.SelectBuilder {
	return Select(column).
		Prefix(`(`).
		Suffix(fmt.Sprintf(`) as %s`, columnName))
}
