package squery

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/spanner"
)

type DataAsEntity map[string]any

func (d DataAsEntity) Data() map[string]any {
	return d
}

type DataAsPrimaryKey []any

func (d DataAsPrimaryKey) PrimaryKey() []any {
	return d
}

type EntityData interface {
	Data() map[string]any
}

type EntityPrimaryKey interface {
	PrimaryKey() []any
}

type Transaction struct {
	conn      *spanner.Client
	mutations []*spanner.Mutation
	mu        sync.Mutex
	deepness  int
}

// MockWrite is used to allow performing Write inside nested operations without any real outcome
func (t *Transaction) MockWrite() *Transaction {
	t.deepness++
	return t
}

func (t *Transaction) Update(table string, e EntityData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.UpdateMap(table, e.Data()))
}

func (t *Transaction) Insert(table string, e EntityData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.InsertMap(table, e.Data()))
}

func (t *Transaction) InsertOrUpdate(table string, e EntityData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.InsertOrUpdateMap(table, e.Data()))
}

func (t *Transaction) Delete(table string, e EntityPrimaryKey) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.Delete(table, spanner.KeySetFromKeys(e.PrimaryKey())))
}

func (t *Transaction) Mutations() []*spanner.Mutation {
	return t.mutations
}

func (t *Transaction) Write(ctx context.Context) (commitTimestamp time.Time, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.deepness > 0 {
		t.deepness--
		return time.Time{}, nil
	}

	if len(t.mutations) == 0 {
		return time.Now(), nil
	}

	return t.conn.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		defer func() {
			t.mutations = t.mutations[:0]
		}()
		return tx.BufferWrite(t.mutations)
	})
}
