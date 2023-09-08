package typesystem

import (
	"reflect"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
)

type Decoration interface {
	GetDecorator() Decorator
}

type DecorationBase struct {
	Decorator Decorator
	Tags      reflect.StructTag
}

func (d DecorationBase) GetDecorator() Decorator {
	return d.Decorator
}

type Decorator interface {
	DecoratorMarker()
	BuildDecoration(tags reflect.StructTag) Decoration
}

type DecoratorBase struct{}

func (db DecoratorBase) DecoratorMarker() {}

type NameDecoration struct {
	DecorationBase

	Name string
}

type Name struct{ DecoratorBase }

func (n Name) BuildDecoration(tags reflect.StructTag) Decoration {
	return NameDecoration{
		DecorationBase: DecorationBase{Decorator: n, Tags: tags},
		Name:           tags.Get("name"),
	}
}

type IImplementsMarker interface {
	ImplementsMarker()
}

type ImplementsInterface[TInterface any] struct{ DecoratorBase }

func (i ImplementsInterface[T]) ImplementsMarker() {}

func (i ImplementsInterface[T]) InterfaceType() typesystem.Type {
	return typesystem.TypeOf(i)
}

type ImplementsDecoration struct {
	DecorationBase

	Interface typesystem.Type
}

func (i ImplementsInterface[TInterface]) BuildDecoration(tags reflect.StructTag) Decoration {
	return ImplementsDecoration{
		DecorationBase: DecorationBase{Decorator: i, Tags: tags},
		Interface:      i.InterfaceType(),
	}
}

func DecorationsForType(typ reflect.Type) []Decoration {
	var decorations []Decoration
	var walkType func(typ reflect.Type)

	walkType = func(typ reflect.Type) {
		for i := 0; i < typ.NumField(); i++ {
			fld := typ.Field(i)

			if fld.Type.Implements(decoratorType) {
				decorator := ExtractDecoratorFromField(fld)

				decorations = append(decorations, decorator.BuildDecoration(fld.Tag))
			} else if fld.Type.Kind() == reflect.Struct && (fld.Anonymous || fld.Name == "_") {
				walkType(fld.Type)
			}
		}
	}

	walkType(typ)

	return decorations
}

func ExtractDecoratorFromField(fld reflect.StructField) Decorator {
	typ := fld.Type

	return reflect.New(typ).Interface().(Decorator)
}

var decoratorType = reflect.TypeOf((*Decorator)(nil)).Elem()
