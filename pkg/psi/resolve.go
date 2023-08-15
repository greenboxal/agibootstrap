package psi

import (
	"context"

	"github.com/pkg/errors"
)

func ResolveChild(parent Path, child PathElement) Path {
	return parent.Child(child)
}

func ResolveEdge[T Node](parent Node, key TypedEdgeKey[T]) (def T) {
	e := parent.GetEdge(key)

	if e == nil {
		return
	}

	return e.To().(T)
}

func MustResolve[T Node](ctx context.Context, g Graph, path Path) (empty T) {
	result, err := Resolve[T](ctx, g, path)

	if err != nil {
		panic(err)
	}

	return result
}

func Resolve[T Node](ctx context.Context, g Graph, path Path) (empty T, _ error) {
	n, err := g.ResolveNode(ctx, path)

	if err != nil {
		return empty, err
	}

	return n.(T), nil
}

func MustResolveChildOrCreate[T Node](ctx context.Context, n Node, name PathElement, factoryFn func() T) (empty T) {
	result, err := ResolveChildOrCreate[T](ctx, n, name, factoryFn)

	if err != nil {
		panic(err)
	}

	return result
}

func ResolveChildOrCreate[T Node](ctx context.Context, n Node, name PathElement, factoryFn func() T) (empty T, _ error) {
	result := n.ResolveChild(ctx, name)

	if result == nil {
		result = factoryFn()

		result.SetParent(n)
	}

	return result.(T), nil
}

func MustResolveOrCreate[T Node](ctx context.Context, g Graph, path Path, factoryFn func() T) (empty T) {
	result, err := ResolveOrCreate[T](ctx, g, path, factoryFn)

	if err != nil {
		panic(err)
	}

	return result
}

func ResolveOrCreate[T Node](ctx context.Context, g Graph, path Path, factoryFn func() T) (empty T, _ error) {
	result, err := g.ResolveNode(ctx, path)

	if err == nil {
		return result.(T), nil
	} else if err != nil && !errors.Is(err, ErrNodeNotFound) {
		return empty, err
	}

	if path.Len() == 0 {
		return empty, ErrNodeNotFound
	}

	parent, err := g.ResolveNode(ctx, path.Parent())

	if err != nil {
		return empty, err
	}

	result = parent.ResolveChild(ctx, path.Name().AsPathElement())

	if result == nil {
		result = factoryFn()

		result.SetParent(parent)
	}

	return result.(T), nil
}

func ResolvePath(ctx context.Context, root Node, path Path) (empty Node, _ error) {
	result := root

	for i, component := range path.components {
		if component.IsEmpty() {
			if i == 0 {
				component.Name = "/"
			} else {
				panic("empty Path component")
			}
		}

		if component.Kind == EdgeKindChild && component.Index == 0 {
			if component.Name == "/" {
				result = root
			} else if component.Name == "." {
				continue
			} else if component.Name == ".." {
				cn := result.Parent()

				if cn != nil {
					result = cn
				}

				continue
			}
		}

		cn := result.ResolveChild(ctx, component)

		if cn == nil {
			break
		}

		result = cn
	}

	if result == nil {
		return empty, ErrNodeNotFound
	}

	return result, nil
}
