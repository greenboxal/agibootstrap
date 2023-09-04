package coreapi

type Checkpoint interface {
	Get() (uint64, bool, error)
	Update(xid uint64, onlyIfGreater bool) error
	Close() error
}
