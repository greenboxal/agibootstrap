package indexing

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ContentAddressableStore interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
	Put(ctx context.Context, value []byte) ([]byte, error)
}

type ChildEntry struct {
	UUID string
}

type PsiIndex interface {
	GetNodeByUUID(uuid string) (psi.Node, error)
	GetEdgesFrom(uuid string) (psi.EdgeIterator, error)
	GetChildrenNodes(parentUuid string) ([]ChildEntry, error)

	AddNode(node psi.Node) error
}
