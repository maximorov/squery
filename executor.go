package squery

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
)

// SqSQLer is an interface for types that can be converted to a SQL query.
type SqSQLer interface {
	ToSql() (string, []interface{}, error)
}

// Entity is an interface for types that can be converted to a slice of any.
type Entity interface {
	ToData() []any
}

// NewExecutor creates a new Executor.
func NewExecutor[E any](conn *spanner.Client) *Executor[E] {
	return &Executor[E]{
		conn: conn,
	}
}

// Executor is a generic struct for executing SQL queries.
type Executor[E any] struct {
	conn *spanner.Client
	dst  E
}

// RowsForStmt executes a query and returns a slice of results.
// If a transaction is provided, it will be used for the query.
func (r *Executor[E]) RowsForStmt(ctx context.Context, stmt spanner.Statement, tx *spanner.ReadWriteTransaction) ([]E, error) {
	var iter *spanner.RowIterator

	if tx != nil {
		iter = tx.Query(ctx, stmt)
	} else {
		iter = r.conn.Single().Query(ctx, stmt)
	}
	defer iter.Stop()

	res := make([]E, 0)
	return res, iter.Do(func(row *spanner.Row) error {
		tmp := &r.dst
		dst := *tmp
		if err := row.Columns(any(&dst).(Entity).ToData()...); err != nil {
			return err
		}
		res = append(res, dst)
		return nil
	})
}

// Rows builds and executes a query, returning a slice of results.
// Additional arguments can be provided as key-value pairs.
func (r *Executor[E]) Rows(ctx context.Context, sb SqSQLer, addArgs ...any) ([]E, error) {
	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	params := make(map[string]any)
	for i, arg := range args {
		i++
		params[fmt.Sprintf(`p%d`, i)] = arg
	}
	if addArgs != nil {
		if len(addArgs)%2 != 0 {
			return nil, fmt.Errorf(`additional arguments must be even`)
		}
		for i := 0; i < len(addArgs); i += 2 {
			key, ok := addArgs[i].(string)
			if !ok {
				return nil, fmt.Errorf(`additional argument key must be a string`)
			}
			params[key] = addArgs[i+1]
		}
	}

	stmt := spanner.Statement{
		SQL:    sql,
		Params: params,
	}

	return r.RowsForStmt(ctx, stmt, nil)
}

// Col executes a query and returns a single column value.
// Additional arguments can be provided as key-value pairs.
func (r *Executor[E]) Col(ctx context.Context, sb SqSQLer, addArgs ...any) (E, error) {
	sql, args, err := sb.ToSql()
	if err != nil {
		return r.dst, err
	}

	params := make(map[string]any)
	for i, arg := range args {
		i++
		params[fmt.Sprintf(`p%d`, i)] = arg
	}
	if addArgs != nil {
		if len(addArgs)%2 != 0 {
			return r.dst, fmt.Errorf(`additional arguments must be even`)
		}
		for i := 0; i < len(addArgs); i += 2 {
			key, ok := addArgs[i].(string)
			if !ok {
				return r.dst, fmt.Errorf(`additional argument key must be a string`)
			}
			params[key] = addArgs[i+1]
		}
	}

	stmt := spanner.Statement{
		SQL:    sql,
		Params: params,
	}

	iter := r.conn.Single().Query(ctx, stmt)
	defer iter.Stop()

	return r.dst, iter.Do(func(row *spanner.Row) error {
		if err := row.Columns(&r.dst); err != nil {
			return err
		}
		return nil
	})
}

// Row builds and executes a query, returning a single row.
func (r *Executor[E]) Row(ctx context.Context, sb SqSQLer) (*E, error) {
	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	params := make(map[string]any)
	for i, arg := range args {
		i++
		params[fmt.Sprintf(`p%d`, i)] = arg
	}

	stmt := spanner.Statement{
		SQL:    sql,
		Params: params,
	}

	return r.RowForStmt(ctx, stmt, nil)
}

// RowForStmt executes a query and returns a single row.
// If a transaction is provided, it will be used for the query.
func (r *Executor[E]) RowForStmt(ctx context.Context, stmt spanner.Statement, tx *spanner.ReadWriteTransaction) (*E, error) {
	ents, err := r.RowsForStmt(ctx, stmt, tx)
	if err != nil {
		return nil, err
	}

	if len(ents) == 0 {
		return nil, spanner.ErrRowNotFound
	}

	return &ents[0], nil
}
