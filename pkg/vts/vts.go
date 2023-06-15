package vts

import "reflect"

type PackageName string

type Package struct {
	Name  PackageName
	Types []*Type
}

type TypeName string

type Type struct {
	Pkg  PackageName
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

func ReflectGoType(typ reflect.Type) *Type {
	// TODO: Implement this by method by converting a reflect.Type to a *Type

	// Convert reflect.Type to *Type
	typeName := TypeName(typ.String())
	return &Type{
		Name: typeName,
	}
}
