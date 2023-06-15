package vts

import "sync"

type PackageName string

type Package struct {
	mu sync.RWMutex

	Name  PackageName
	Types []*Type
}

func (p *Package) ResolveType(name string) *Type {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, typ := range p.Types {
		if typ.Name.Name == name {
			return typ
		}
	}

	return nil
}

type TypeName struct {
	Pkg  PackageName
	Name string
}

type Type struct {
	Name TypeName

	Members []TypeMember
}

type TypeMember interface {
	GetName() string
	GetDeclarationType() TypeName
}

type Method struct {
	DeclarationType TypeName
	Name            string

	Parameters []Parameter
	Results    []Parameter

	TypeParameters []Parameter
}

func (m *Method) GetName() string              { return m.Name }
func (m *Method) GetDeclarationType() TypeName { return m.DeclarationType }

type Parameter struct {
	Name string
	Type TypeName
}

type Field struct {
	DeclarationType TypeName
	Name            string
	Type            TypeName
}

func (f *Field) GetName() string              { return f.Name }
func (f *Field) GetDeclarationType() TypeName { return f.DeclarationType }
