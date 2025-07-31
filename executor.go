package squery

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
)

type SqSQLer interface {
	ToSql() (string, []interface{}, error)
}

type Entity interface {
	ToData() []any
}

func NewExecutor[E any](conn *spanner.Client, dst E) *Executor[E] {
	return &Executor[E]{
		conn: conn,
		dst:  dst,
	}
}

type Executor[E any] struct {
	conn *spanner.Client
	dst  E
}

func (r *Executor[E]) Iterate(ctx context.Context, stmt spanner.Statement, tx *spanner.ReadWriteTransaction) ([]E, error) {
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

func (r *Executor[E]) IterateAfterSq(ctx context.Context, sb SqSQLer, addArgs ...any) ([]E, error) {
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

	return r.Iterate(ctx, stmt, nil)
}

func (r *Executor[E]) GetSingleRow(ctx context.Context, sb SqSQLer, addArgs ...any) (E, error) {
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

func (r *Executor[E]) RowAfterSq(ctx context.Context, sb SqSQLer) (*E, error) {
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

	return r.Row(ctx, stmt, nil)
}

func (r *Executor[E]) Row(ctx context.Context, stmt spanner.Statement, tx *spanner.ReadWriteTransaction) (*E, error) {
	ents, err := r.Iterate(ctx, stmt, tx)
	if err != nil {
		return nil, err
	}

	if len(ents) == 0 {
		return nil, spanner.ErrRowNotFound
	}

	return &ents[0], nil
}
