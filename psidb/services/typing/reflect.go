package typing

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Registry struct {
	Root psi.Path `json:"root"`
}

func (r *Registry) ResolveInterfacePath(name string) psi.Path {
	return r.Root.Child(psi.PathElement{Name: name})
}

func (r *Registry) ResolveInterface(ctx context.Context, iface *psi.VTable) (*Interface, error) {
	tx := coreapi.GetTransaction(ctx)

	return psi.ResolveOrCreate[*Interface](
		ctx,
		tx.Graph(),
		r.ResolveInterfacePath(iface.Name()),
		func() *Interface {
			return BuildNodeInterfaceType(iface.Interface())
		},
	)
}

func (r *Registry) ResolveNodeType(ctx context.Context, typ psi.NodeType) (*Type, error) {
	tx := coreapi.GetTransaction(ctx)

	return psi.ResolveOrCreate[*Type](
		ctx,
		tx.Graph(),
		r.ResolveInterfacePath(typ.Name()),
		func() *Type {
			t := NewType(typ.Name())

			for _, iface := range typ.Interfaces() {
				i, err := r.ResolveInterface(ctx, iface)

				if err != nil {
					panic(err)
				}

				t.SetEdge(EdgeKindImplements.Named(iface.Name()), i)
			}

			return t
		},
	)
}

func BuildNodeInterfaceType(ni psi.NodeInterface) *Interface {
	iface := NewInterface(ni.Name())

	//for _, action := range ni.Actions() {
	//requestType := featureextractors.ReflectSchemaForType(action.RequestType.RuntimeType())
	//responseType := featureextractors.ReflectSchemaForType(action.ResponseType.RuntimeType())
	/*m := NewMethod(action.Name, requestType, responseType)

	m.SetParent(iface)*/
	//}

	return iface
}
