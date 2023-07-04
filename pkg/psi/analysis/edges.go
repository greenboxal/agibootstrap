package analysis

import "github.com/greenboxal/agibootstrap/pkg/psi"

const EdgeKindSymbol = psi.TypedEdgeKind[*Symbol]("codex.analysis.symbol")
const EdgeKindScope = psi.TypedEdgeKind[*Scope]("codex.analysis.scope")
const EdgeKindDefinition = psi.TypedEdgeKind[psi.Node]("codex.analysis.definition")
const EdgeKindReference = psi.TypedEdgeKind[psi.Node]("codex.analysis.reference")

type ScopedNode interface {
	psi.Node

	PsiNodeScope() *Scope
}

type DeclarationNode interface {
	ScopedNode

	PsiNodeDeclaration() *Symbol
}

func GetNodeSymbol(node psi.Node) *Symbol {
	if decl, ok := node.(DeclarationNode); ok {
		return decl.PsiNodeDeclaration()
	}

	sym, _ := psi.GetEdge[*Symbol](node, EdgeKindSymbol.Singleton())

	return sym
}

func GetNodeScope(node psi.Node) (s *Scope) {
	for node != nil {
		s, _ = psi.GetEdge[*Scope](node, EdgeKindScope.Singleton())

		if s != nil {
			break
		}

		node = node.Parent()
	}

	return
}
