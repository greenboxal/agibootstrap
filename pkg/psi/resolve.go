package psi

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

func Resolve(root Node, path string) (Node, error) {
	p, err := ParsePath(path)

	if err != nil {
		return nil, err
	}

	return ResolvePath(root, p)
}

func ResolvePath(root Node, path Path) (Node, error) {
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

		cn := result.ResolveChild(component)

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
