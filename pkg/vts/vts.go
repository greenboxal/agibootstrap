package vts

import (
	"fmt"
	"sync"
)

type PackageName string

type Package struct {
	mu sync.RWMutex

	Name  PackageName
	Types []*Type
}

// ResolveType resolves a type by name within the package.
// It searches for a type with a matching name in the package's list of types.
// If a matching type is found, it returns a pointer to the type.
// Otherwise, it returns nil.
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

func (p *Package) String() string {
	return string(p.Name)
}

type TypeName struct {
	Pkg  PackageName
	Name string
}

func (tn *TypeName) String() string {
	return fmt.Sprintf("%s/%s", tn.Pkg, tn.Name)
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

func (m *Method) String() string {
	return fmt.Sprintf("%s.%s", m.DeclarationType, m.Name)
}

func (m *Method) GetName() string              { return m.Name }
func (m *Method) GetDeclarationType() TypeName { return m.DeclarationType }

type Parameter struct {
	Name string
	Type TypeName
}

// Field represents a field in a type.
type Field struct {
	// DeclarationType represents the type of the field declaration.
	DeclarationType TypeName

	// Name represents the name of the field.
	Name string

	// Type represents the type of the field.
	Type TypeName
}

func (f *Field) GetName() string { return f.Name }

func (f *Field) GetDeclarationType() TypeName { return f.DeclarationType }

type Scope struct {
	Packages map[PackageName]*Package
	Types    map[TypeName]*Type
}

func (s *Scope) AddPackage(pkg *Package) {
	s.Packages[pkg.Name] = pkg

	for _, typ := range pkg.Types {
		s.Types[typ.Name] = typ
	}
}

// NewScope returns a new instance of Scope.
// It creates an empty scope with initialized maps for Packages and Types.
func NewScope() *Scope {
	return &Scope{
		Packages: make(map[PackageName]*Package),
		Types:    map[TypeName]*Type{},
	}
}
