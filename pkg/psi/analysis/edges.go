package analysis

import "github.com/greenboxal/agibootstrap/pkg/psi"

const EdgeKindSymbol = psi.TypedEdgeKind[*Symbol]("codex.analysis.symbol")
const EdgeKindScope = psi.TypedEdgeKind[*Scope]("codex.analysis.scope")
const EdgeKindDefinition = psi.TypedEdgeKind[psi.Node]("codex.analysis.definition")
const EdgeKindReference = psi.TypedEdgeKind[psi.Node]("codex.analysis.reference")
const EdgeKindImport = psi.TypedEdgeKind[*Scope]("codex.analysis.import")

type ScopedNode interface {
	psi.Node

	PsiNodeScope() *Scope
}

type DeclarationNode interface {
	ScopedNode

	PsiNodeDeclaration() *Symbol
}

func DefineNodeSymbol(node psi.Node, name string) *Symbol {
	if sym := GetNodeSymbol(node); sym != nil && sym.Name == name {
		return sym
	}

	scope := GetNodeScope(node)
	sym := scope.GetOrCreateSymbol(name)

	SetNodeSymbol(node, sym)

	return sym
}

func SetNodeSymbol(node psi.Node, sym *Symbol) {
	sym.UpdateDefinition(node)
}

func GetNodeSymbol(node psi.Node) *Symbol {
	if decl, ok := node.(DeclarationNode); ok {
		return decl.PsiNodeDeclaration()
	}

	sym, _ := psi.GetEdge[*Symbol](node, EdgeKindSymbol.Singleton())

	return sym
}

func SetNodeScope(node psi.Node, scope *Scope) {
	psi.UpdateEdge(node, EdgeKindScope.Singleton(), scope)
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
