package squery

import (
	"cloud.google.com/go/spanner"
)

type TransactionFactory struct {
	conn *spanner.Client
}

func NewTransactionFactory(conn *spanner.Client) *TransactionFactory {
	return &TransactionFactory{
		conn: conn,
	}
}

func (f *TransactionFactory) NewTransaction() *Transaction {
	return &Transaction{
		conn: f.conn,
	}
}

func (f *TransactionFactory) NewTransactionOrMock(tx *Transaction) *Transaction {
	if tx != nil {
		return tx.MockWrite()
	}

	return f.NewTransaction()
}
