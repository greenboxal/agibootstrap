package indexing

import (
	"context"

	"github.com/google/uuid"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Scope2 struct {
	psi.NodeBase

	IndexName string `json:"index_name"`

	IndexManager *Manager     `inject:"" json:"-"`
	Embedder     NodeEmbedder `inject:"" json:"-"`

	Index NodeIndex `json:"-"`
}

var ScopeType = psi.DefineNodeType[*Scope2]()
var ScopeEdge = psi.DefineEdgeType[*Scope2]("psidb.scope2")
var ScopeRootEdge = psi.DefineEdgeType[psi.Node]("psidb.scope.root2")

func NewScope() *Scope2 {
	scp := &Scope2{}
	scp.IndexName = uuid.NewString()
	scp.Init(scp, psi.WithNodeType(ScopeType))

	return scp
}

func (scp *Scope2) PsiNodeName() string { return scp.IndexName }

func (scp *Scope2) SetRoot(root psi.Node) {
	scp.SetEdge(ScopeRootEdge.Singleton(), root)
}

func (scp *Scope2) GetRoot() psi.Node {
	return psi.GetEdgeOrNil[psi.Node](scp, ScopeRootEdge.Singleton())
}

func (scp *Scope2) Upsert(ctx context.Context, node psi.Node) error {
	idx, err := scp.GetIndex(ctx)

	if err != nil {
		return err
	}

	return idx.IndexNode(ctx, node)
}

func (scp *Scope2) GetIndex(ctx context.Context) (NodeIndex, error) {
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

func (scp *Scope2) Close() error {
	if scp.Index != nil {
		if err := scp.Index.Close(); err != nil {
			return err
		}

		scp.Index = nil
	}

	return nil
}

func GetNodeScope(node psi.Node) *Scope2 {
	return psi.GetEdgeOrNil[*Scope2](node, ScopeEdge.Singleton())
}

func SetNodeScope(node psi.Node, scp *Scope2) {
	node.SetEdge(ScopeEdge.Singleton(), scp)
}

func GetHierarchyScope(ctx context.Context, node psi.Node) *Scope2 {
	for ; node != nil; node = node.Parent() {
		scp := node.ResolveChild(ctx, ScopeEdge.Singleton().AsPathElement())

		if scp != nil {
			scp, ok := scp.(*Scope2)

			if !ok {
				node.ResolveChild(ctx, ScopeEdge.Singleton().AsPathElement())
				panic("scope edge is not a scope")
			}

			return scp
		}
	}

	return nil
}
