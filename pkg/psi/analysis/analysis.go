package analysis

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Manager struct {
}

func NewScope() *Scope {
	s := &Scope{}

	s.Init(s)

	return s
}

type Scope struct {
	psi.NodeBase
}

func (s *Scope) ParentScope() *Scope {
	p, ok := s.Parent().(*Scope)

	if !ok {
		return nil
	}

	return p
}

func (s *Scope) GetSymbol(name string) *Symbol {
	key := EdgeKindSymbol.Named(name)
	sym, _ := psi.GetEdge[*Symbol](s, key)

	return sym
}

func (s *Scope) GetOrCreateSymbol(name string) *Symbol {
	key := EdgeKindSymbol.Named(name)

	return psi.GetOrCreateEdge(s, key, func() *Symbol {
		symbol := NewSymbol()
		symbol.Name = name

		s.AddChildNode(symbol)

		return symbol
	})
}

func (s *Scope) RemoveSymbol(symbol *Symbol) {
	s.RemoveChildNode(symbol)
}

func NewSymbol() *Symbol {
	s := &Symbol{}

	s.Init(s)

	return s
}

type Symbol struct {
	psi.NodeBase `json:"-"`

	Name string `json:"name"`
}

func (s *Symbol) PsiNodeName() string { return s.Name }

func (s *Symbol) Scope() *Scope {
	p, ok := s.Parent().(*Scope)

	if !ok {
		return nil
	}

	return p
}

func (s *Symbol) UpdateDefinition(node psi.Node) {
	psi.UpdateEdge(s, EdgeKindDefinition.Singleton(), node)
}
