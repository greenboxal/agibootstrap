package analysis

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func NewSymbol() *Symbol {
	s := &Symbol{}

	s.Init(s, psi.WithNodeType(SymbolType))

	return s
}

type Symbol struct {
	psi.NodeBase

	Name string `json:"name"`

	RootDistance      int `json:"root_distance"`
	ScopeDistance     int `json:"scope_distance"`
	ReferenceDistance int `json:"reference_distance"`

	IsLocal    bool `json:"is_local"`
	IsResolved bool `json:"is_resolved"`
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
	s.IsResolved = node != nil

	psi.UpdateEdge(s, EdgeKindDefinition.Singleton(), node)
	psi.UpdateEdge(node, EdgeKindSymbol.Singleton(), s)

	s.Invalidate()
}

func (s *Symbol) Resolve(ctx context.Context) (bool, error) {
	if s.IsResolved || s.IsLocal {
		if s.IsLocal {
			s.ReferenceDistance = 0
		}

		return true, nil
	}

	scope := s.Scope()

	if scope == nil {
		return false, nil
	}

	resolved, d, err := scope.Resolve(ctx, s.Name)

	if err != nil {
		return false, err
	}

	if resolved == nil || (resolved == s && !s.IsResolved) {
		return false, nil
	}

	s.ReferenceDistance = d

	s.UpdateDefinition(resolved)

	return s.IsResolved, nil
}
