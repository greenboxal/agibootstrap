package indexing

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Scope struct {
	psi.NodeBase
}

var ScopeType = psi.DefineNodeType[*Scope]()
var ScopeEdge = psi.DefineEdgeType[*Scope]("psidb.scope")
var ScopeRootEdge = psi.DefineEdgeType[psi.Node]("psidb.scope.root")

func NewScope() *Scope {
	scp := &Scope{}
	scp.Init(scp, psi.WithNodeType(ScopeType))

	return scp
}

func (scp *Scope) SetRoot(root psi.Node) {
	scp.SetEdge(ScopeRootEdge.Singleton(), root)
}

func (scp *Scope) GetRoot() psi.Node {
	return psi.GetEdgeOrNil[psi.Node](scp, ScopeRootEdge.Singleton())
}

func (scp *Scope) Upsert(ctx context.Context, node psi.Node) error {

	return nil
}

func GetNodeScope(node psi.Node) *Scope {
	return psi.GetEdgeOrNil[*Scope](node, ScopeEdge.Singleton())
}

func SetNodeScope(node psi.Node, scp *Scope) {
	node.SetEdge(ScopeEdge.Singleton(), scp)
}

func GetHierarchyScope(node psi.Node) *Scope {
	for ; node != nil; node = node.Parent() {
		scp := GetNodeScope(node)

		if scp != nil {
			return scp
		}
	}

	return nil
}
