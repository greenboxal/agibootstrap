package psi

type NodeIterator interface {
	Value() Node
	Node() Node
	Next() bool
}

type nodeSliceIterator struct {
	current Node
	items   []Node
}

func (n *nodeSliceIterator) Value() Node { return n.Node() }

func (n *nodeSliceIterator) Node() Node {
	return n.current
}

func (n *nodeSliceIterator) Next() bool {
	if len(n.items) == 0 {
		return false
	}

	n.current = n.items[0]
	n.items = n.items[1:]

	return true
}

func (n *nodeSliceIterator) Prepend(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{iterator, n}}
}

func (n *nodeSliceIterator) Append(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{n, iterator}}
}

type nodeChildrenIterator struct {
	parent  *NodeBase
	current Node
	index   int
}

func (n *nodeChildrenIterator) Value() Node { return n.Node() }

func (n *nodeChildrenIterator) Node() Node {
	return n.current
}

func (n *nodeChildrenIterator) Next() bool {
	if n.index >= n.parent.children.Len() {
		return false
	}

	n.current = n.parent.children.Get(n.index)
	n.index++

	return true
}

func (n *nodeChildrenIterator) Prepend(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{iterator, n}}
}

func (n *nodeChildrenIterator) Append(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{n, iterator}}
}

type nestedNodeIterator struct {
	current   NodeIterator
	iterators []NodeIterator
}

func (n *nestedNodeIterator) Value() Node { return n.Node() }

func (n *nestedNodeIterator) Node() Node {
	if n.current == nil {
		return nil
	}

	return n.current.Node()
}

func (n *nestedNodeIterator) Next() bool {
	for n.current == nil || !n.current.Next() {
		if len(n.iterators) == 0 {
			return false
		}

		n.current = n.iterators[0]
		n.iterators = n.iterators[1:]
	}

	return true
}

func (n *nestedNodeIterator) Prepend(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{iterator, n}}
}

func (n *nestedNodeIterator) Append(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{n, iterator}}
}

func AppendNodeIterator(iterators ...NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: iterators}
}
