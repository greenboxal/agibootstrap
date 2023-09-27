package codeanalysis

import (
	"net/url"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

func DefineNodeScope(node psi.Node) *Scope {
	scp := psi.GetEdgeOrNil[*Scope](node, ScopeEdge.Singleton())

	if scp == nil {
		parent := GetNodeHierarchyScope(node.Parent())
		name := url.PathEscape(node.SelfIdentity().String())

		if parent != nil {
			scp = parent.GetChildScope(name, true)
		} else {
			scp = NewScope(nil, name)
		}

		node.SetEdge(ScopeEdge.Singleton(), scp)
	}

	return scp
}

func GetNodeHierarchyScope(node psi.Node) *Scope {
	for n := node; n != nil; n = n.Parent() {
		scp := psi.GetEdgeOrNil[*Scope](n, ScopeEdge.Singleton())

		if scp != nil {
			return scp
		}
	}

	return nil
}

type Scope struct {
	psi.NodeBase

	Name string `json:"name"`
}

func NewScope(parent *Scope, name string) *Scope {
	scp := &Scope{
		Name: name,
	}

	scp.Init(scp)

	if parent != nil {
		scp.SetEdge(ScopeEdge.Singleton(), parent)
	}

	return scp
}

func (s *Scope) PsiNodeName() string { return s.Name }

func (s *Scope) ParentScope() *Scope {
	return psi.GetEdgeOrNil[*Scope](s, ScopeEdge.Singleton())
}

func (s *Scope) GetSymbol(name string) *Symbol {
	return psi.GetEdgeOrNil[*Symbol](s, SymbolEdge.Named(name))
}

func (s *Scope) GetSymbols() []*Symbol {
	return psi.GetEdges(s, SymbolEdge)
}

func (s *Scope) AddSymbol(sym *Symbol) {
	s.SetEdge(SymbolEdge.Named(sym.Name), sym)
}

func (s *Scope) DefineSymbol(name string) *Symbol {
	if sym := s.GetSymbol(name); sym != nil {
		return sym
	}

	sym := NewSymbol(name)
	sym.SetParent(s)
	s.AddSymbol(sym)

	return sym
}

func (s *Scope) ResolveSymbol(names ...string) *Symbol {
	scp := s

	for _, name := range names {
		sym := scp.GetSymbol(name)

		if sym == nil {
			if ps := scp.ParentScope(); ps != nil {
				sym = ps.ResolveSymbol(name)
			}

			if sym == nil {
				return nil
			}
		}

		decl := sym.GetDeclaration()

		if decl != nil {
			scp = GetNodeHierarchyScope(decl)
		} else {
			return nil
		}
	}

	return nil
}

func (s *Scope) GetChildScope(name string, create bool) *Scope {
	if scp := psi.GetEdgeOrNil[*Scope](s, ScopeEdge.Named(name)); scp != nil {
		return scp
	}

	if create {
		scp := NewScope(s, name)
		s.SetParent(scp)
		return scp
	}

	return nil
}

var ScopeType = psi.DefineNodeType[*Scope]()
var ScopeEdge = psi.DefineEdgeType[*Scope]("ca.Scope")
