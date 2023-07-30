package graphfs

import (
	"context"
)

// SuperBlockOperations is the interface for superblock operations.
type SuperBlockOperations interface {
	// GetRoot returns the root dentry of the superblock.
	GetRoot(ctx context.Context) (*CacheEntry, error)
}

type SuperBlock interface {
	SuperBlockOperations

	// UUID returns the UUID of the superblock.
	UUID() string

	// GetSuperBlockBase returns the superblock base.
	GetSuperBlockBase() *SuperBlockBase
	// GetSuperBlock returns the superblock.
	GetSuperBlock() SuperBlock

	INodeOperations() INodeOperations
	NodeHandleOperations() NodeHandleOperations

	Close(ctx context.Context) error
}

type SuperBlockBase struct {
	SuperBlock

	uuid string

	root *CacheEntry

	iops INodeOperations
	nops NodeHandleOperations
}

func (s *SuperBlockBase) UUID() string                               { return s.uuid }
func (s *SuperBlockBase) GetSuperBlockBase() *SuperBlockBase         { return s }
func (s *SuperBlockBase) GetSuperBlock() SuperBlock                  { return s.SuperBlock }
func (s *SuperBlockBase) INodeOperations() INodeOperations           { return s.iops }
func (s *SuperBlockBase) NodeHandleOperations() NodeHandleOperations { return s.nops }

func (s *SuperBlockBase) Init(self SuperBlock, uuid string, iops INodeOperations, nops NodeHandleOperations) {
	if s.SuperBlock != nil {
		panic("SuperBlockBase already initialized")
	}

	s.SuperBlock = self
	s.uuid = uuid
	s.iops = iops
	s.nops = nops
}

func (s *SuperBlockBase) Close() error {
	return nil
}
