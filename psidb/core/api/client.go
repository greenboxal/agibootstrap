package coreapi

type Client interface {
	TransactionOperations

	Close() error
}
