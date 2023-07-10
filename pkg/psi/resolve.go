package psi

import "context"

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

func Resolve(ctx context.Context, root Node, path string) (Node, error) {
	p, err := ParsePath(path)

	if err != nil {
		return nil, err
	}

	return ResolvePath(ctx, root, p)
}

func ResolvePath(ctx context.Context, root Node, path Path) (Node, error) {
	rootPath := root.CanonicalPath()

	if !path.IsRelative() {
		rel, err := path.RelativeTo(rootPath)

		if err != nil {
			return nil, err
		}

		path = rel
	}

	if root == nil {
		return nil, ErrNodeNotFound
	}

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
		return nil, ErrNodeNotFound
	}

	return result, nil
}
