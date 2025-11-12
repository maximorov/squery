package squery

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/spanner"
)

// DataAsEntity is a map of column names to values, representing a Spanner entity.
type DataAsEntity map[string]any

// Data returns the underlying map.
func (d DataAsEntity) Data() map[string]any {
	return d
}

// DataAsPrimaryKey is a slice of values representing a Spanner primary key.
type DataAsPrimaryKey []any

// PrimaryKey returns the underlying slice.
func (d DataAsPrimaryKey) PrimaryKey() []any {
	return d
}

// EntityData is an interface for types that can provide their data as a map.
type EntityData interface {
	Data() map[string]any
}

// EntityPrimaryKey is an interface for types that can provide their primary key as a slice of values.
type EntityPrimaryKey interface {
	PrimaryKey() []any
}

// Transaction provides a way to build and execute a set of mutations.
type Transaction struct {
	conn      *spanner.Client
	mutations []*spanner.Mutation
	mu        sync.Mutex
	deepness  int
}

// MockWrite is used to allow performing Write inside nested operations without any real outcome.
// It increments the deepness counter.
func (t *Transaction) MockWrite() *Transaction {
	t.deepness++
	return t
}

// Update adds an update mutation to the transaction.
func (t *Transaction) Update(table string, e EntityData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.UpdateMap(table, e.Data()))
}

// Insert adds an insert mutation to the transaction.
func (t *Transaction) Insert(table string, e EntityData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.InsertMap(table, e.Data()))
}

// InsertOrUpdate adds an insert-or-update mutation to the transaction.
func (t *Transaction) InsertOrUpdate(table string, e EntityData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.InsertOrUpdateMap(table, e.Data()))
}

// Delete adds a delete mutation to the transaction.
func (t *Transaction) Delete(table string, e EntityPrimaryKey) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mutations = append(t.mutations, spanner.Delete(table, spanner.KeySetFromKeys(e.PrimaryKey())))
}

// Mutations returns the list of mutations in the transaction.
func (t *Transaction) Mutations() []*spanner.Mutation {
	return t.mutations
}

// Write executes the buffered mutations in a read-write transaction.
// If the transaction is mocked (deepness > 0), it does nothing.
// If there are no mutations, it returns the current time.
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
