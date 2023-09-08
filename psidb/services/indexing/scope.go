package indexing

import (
	"context"

	"github.com/google/uuid"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Scope struct {
	psi.NodeBase

	IndexName string `json:"index_name"`

	IndexManager *Manager     `inject:"" json:"-"`
	Embedder     NodeEmbedder `inject:"" json:"-"`

	Index NodeIndex `json:"-"`
}

var ScopeType = psi.DefineNodeType[*Scope]()
var ScopeEdge = psi.DefineEdgeType[*Scope]("psidb.scope")
var ScopeRootEdge = psi.DefineEdgeType[psi.Node]("psidb.scope.root")

func NewScope() *Scope {
	scp := &Scope{}
	scp.IndexName = uuid.NewString()
	scp.Init(scp, psi.WithNodeType(ScopeType))

	return scp
}

func (scp *Scope) PsiNodeName() string { return scp.IndexName }

func (scp *Scope) SetRoot(root psi.Node) {
	scp.SetEdge(ScopeRootEdge.Singleton(), root)
}

func (scp *Scope) GetRoot() psi.Node {
	return psi.GetEdgeOrNil[psi.Node](scp, ScopeRootEdge.Singleton())
}

func (scp *Scope) Upsert(ctx context.Context, node psi.Node) error {
	idx, err := scp.GetIndex(ctx)

	if err != nil {
		return err
	}

	return idx.IndexNode(ctx, node)
}

func (scp *Scope) GetIndex(ctx context.Context) (NodeIndex, error) {
	if scp.IndexManager == nil {
		tx := coreapi.GetTransaction(ctx)

		scp.IndexManager = inject.Inject[*Manager](tx.Graph().Services())
	}

	if scp.Embedder == nil {
		tx := coreapi.GetTransaction(ctx)

		scp.Embedder = inject.Inject[NodeEmbedder](tx.Graph().Services())
	}

	if scp.Index == nil {
		idx, err := scp.IndexManager.OpenNodeIndex(ctx, scp.IndexName, scp.Embedder)

		if err != nil {
			return nil, err
		}

		scp.Index = idx
	}

	return scp.Index, nil
}

func (scp *Scope) Close() error {
	if scp.Index != nil {
		if err := scp.Index.Close(); err != nil {
			return err
		}

		scp.Index = nil
	}

	return nil
}

func GetNodeScope(node psi.Node) *Scope {
	return psi.GetEdgeOrNil[*Scope](node, ScopeEdge.Singleton())
}

func SetNodeScope(node psi.Node, scp *Scope) {
	node.SetEdge(ScopeEdge.Singleton(), scp)
}

func GetHierarchyScope(ctx context.Context, node psi.Node) *Scope {
	for ; node != nil; node = node.Parent() {
		scp := node.ResolveChild(ctx, ScopeEdge.Singleton().AsPathElement())

		if scp != nil {
			scp, ok := scp.(*Scope)

			if !ok {
				node.ResolveChild(ctx, ScopeEdge.Singleton().AsPathElement())
				panic("scope edge is not a scope")
			}

			return scp
		}
	}

	return nil
}
