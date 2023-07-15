package analysis

import "github.com/greenboxal/agibootstrap/pkg/psi"

var ScopeType = psi.DefineNodeType[*Scope]()
var SymbolType = psi.DefineNodeType[*Symbol]()

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

func DefineNodeScope(node psi.Node) *Scope {
	scope := GetDirectNodeScope(node)

	if scope == nil {
		scope = NewScope()

		if p := node.Parent(); p != nil {
			if parentScope := GetEffectiveNodeScope(p); parentScope != nil {
				scope.SetParent(parentScope)
			}
		} else {
			scope.SetParent(node)
		}

		SetNodeScope(node, scope)
	}

	return scope
}

func DefineNodeSymbol(node psi.Node, name string) *Symbol {
	if node.Parent() == nil {
		panic("cannot define symbol without parent scope")
	}

	scope := GetEffectiveNodeScope(node.Parent())

	if scope == nil {
		panic("cannot define symbol without parent scope")
	}

	return scope.GetOrCreateSymbol(name)
}

func SetNodeScope(node psi.Node, scope *Scope) {
	psi.UpdateEdge(node, EdgeKindScope.Singleton(), scope)
}

func GetDirectNodeScope(node psi.Node) (s *Scope) {
	s, _ = psi.GetEdge[*Scope](node, EdgeKindScope.Singleton())

	return s
}

func GetEffectiveNodeScope(node psi.Node) (s *Scope) {
	for node != nil {
		s = GetDirectNodeScope(node)

		if s != nil {
			break
		}

		node = node.Parent()
	}

	return
}
