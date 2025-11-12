package squery

import (
	"cloud.google.com/go/spanner"
)

// TransactionFactory is a factory for creating Transaction objects.
type TransactionFactory struct {
	conn *spanner.Client
}

// NewTransactionFactory creates a new TransactionFactory.
func NewTransactionFactory(conn *spanner.Client) *TransactionFactory {
	return &TransactionFactory{
		conn: conn,
	}
}

// NewTransaction creates a new Transaction.
func (f *TransactionFactory) NewTransaction() *Transaction {
	return &Transaction{
		conn: f.conn,
	}
}

// NewTransactionOrMock returns a new transaction or a mock if a transaction is already in progress.
// This is useful for nested operations that should be part of the same transaction.
func (f *TransactionFactory) NewTransactionOrMock(tx *Transaction) *Transaction {
	if tx != nil {
		return tx.MockWrite()
	}

	return f.NewTransaction()
}
