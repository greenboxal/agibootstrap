package analysis

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func NewScope(root psi.Node) *Scope {
	s := &Scope{
		root: root,
	}

	s.Init(s, psi.WithNodeType(ScopeType))

	return s
}

type Scope struct {
	psi.NodeBase

	root psi.Node
}

func (s *Scope) NewChildScope(root psi.Node) *Scope {
	child := NewScope(root)

	s.AddChildNode(child)

	return child
}

func (s *Scope) ParentScope() *Scope {
	p, ok := s.Parent().(*Scope)

	if !ok {
		return nil
	}

	return p
}

func (s *Scope) Resolve(ctx context.Context, name string) (sym *Symbol, d int, err error) {
	if sym = s.resolveLocal(name); sym != nil {
		d = 0
		return
	}

	if sym, d = s.resolveImport(name); sym != nil {
		return
	}

	if p, ok := s.Parent().(*Scope); ok {
		return p.Resolve(ctx, name)
	}

	return nil, -1, nil
}

func (s *Scope) resolveImport(name string) (*Symbol, int) {
	for _, imp := range s.Imports() {
		if sym := imp.resolveLocal(name); sym != nil {
			return sym, 0
		}
	}

	return nil, -1
}

func (s *Scope) resolveLocal(name string) *Symbol {
	key := EdgeKindSymbol.Named(name)
	sym, _ := psi.GetEdge[*Symbol](s, key)

	if sym == nil {
		return nil
	}

	if !sym.IsResolved {
		return nil
	}

	return sym
}

func (s *Scope) Imports() []*Scope {
	filtered := iterators.Filter(s.Edges(), func(t psi.Edge) bool {
		return t.Kind() == EdgeKindImport.Kind()
	})

	scopes := iterators.Map(filtered, func(t psi.Edge) *Scope {
		return t.To().(*Scope)
	})

	return iterators.ToSlice(scopes)
}

func (s *Scope) ImportScope(imp *Scope) {
	psi.UpdateEdge(s, EdgeKindImport.Named(imp.CanonicalPath().String()), imp)
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
