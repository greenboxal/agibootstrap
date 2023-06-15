package golang

import (
	"go/types"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/vts"
)

type TypeSystemProvider struct {
	mu   sync.RWMutex
	pkgs map[vts.PackageName]*vts.Package
	typs map[vts.TypeName]*vts.Type
}

func NewTypeSystemProvider() *TypeSystemProvider {
	return &TypeSystemProvider{
		pkgs: make(map[vts.PackageName]*vts.Package),
		typs: make(map[vts.TypeName]*vts.Type),
	}
}

func (tsp *TypeSystemProvider) IntrospectPackage(t *types.Package) *vts.Package {
	tsp.mu.Lock()
	defer tsp.mu.Unlock()

	name := vts.PackageName(t.Name())

	if tsp.pkgs[name] != nil {
		return tsp.pkgs[name]
	}

	pkg := &vts.Package{
		Name: vts.PackageName(t.Name()),
	}

	tsp.pkgs[name] = pkg

	return pkg
}

func (tsp *TypeSystemProvider) IntrospectType(t types.Type) *vts.Type {
	// TODO: Implement this correctly by caching the result by name in a map.
	tsp.mu.Lock()
	defer tsp.mu.Unlock()

	// Check if the type is already in the cache
	if typ, ok := tsp.typs[vts.TypeName{
		Pkg:  vts.PackageName(t.(*types.Named).Obj().Pkg().Name()),
		Name: t.(*types.Named).Obj().Name(),
	}]; ok {
		return typ
	}

	// Create a new type and add it to the cache
	typ := &vts.Type{
		Name: vts.TypeName{
			Pkg:  vts.PackageName(t.(*types.Named).Obj().Pkg().Name()),
			Name: t.(*types.Named).Obj().Name(),
		},
	}

	tsp.typs[typ.Name] = typ

	return typ
}

func (tsp *TypeSystemProvider) ResolvePackage(name vts.PackageName) *vts.Package {
	return tsp.pkgs[name]
}

func (tsp *TypeSystemProvider) ResolveType(name vts.TypeName) *vts.Type {
	pkg := tsp.ResolvePackage(name.Pkg)

	return pkg.ResolveType(name.Name)
}
