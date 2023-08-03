package psi

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

func (n *NodeBase) Children() []Node                                 { return iterators.ToSlice[Node](n.children.Iterator()) }
func (n *NodeBase) ChildrenList() collectionsfx.ObservableList[Node] { return &n.children }
func (n *NodeBase) ChildrenIterator() NodeIterator                   { return n.children.Iterator() }

func (n *NodeBase) Parent() Node                                { return n.parent.Value() }
func (n *NodeBase) ParentProperty() obsfx.ObservableValue[Node] { return &n.parent }

func (n *NodeBase) SetParent(parent Node) {
	if n == parent || n.self == parent || (parent != nil && n == parent.PsiNodeBase()) {
		panic("invalid parent (cycle)")
	}

	n.parent.SetValue(parent)
}

func (n *NodeBase) IndexOfChild(node Node) int {
	return n.children.IndexOf(node)
}

func (n *NodeBase) PreviousSibling() Node {
	if n.Parent() == nil {
		return nil
	}

	p := n.Parent().PsiNodeBase()
	idx := p.children.IndexOf(n.self)

	if idx <= 0 {
		return nil
	}

	return p.children.Get(idx - 1)
}

func (n *NodeBase) NextSibling() Node {
	if n.Parent() == nil {
		return nil
	}

	p := n.Parent().PsiNodeBase()
	idx := p.children.IndexOf(n.self)

	if idx == -1 || idx >= p.children.Len()-1 {
		return nil
	}

	return p.children.Get(idx + 1)
}

// AddChildNode adds a child node to the current node.
// If the child node is already a child of the current node, no action is taken.
// The child node is appended to the list of children nodes of the current node.
// Then, the child node is attached to the same graph as the parent node.
//
// Parameters:
// - child: The child node to be added.
func (n *NodeBase) AddChildNode(child Node) {
	existingIdx := n.children.IndexOf(child)

	if existingIdx != -1 {
		return
	}

	n.children.Add(child)
}

// RemoveChildNode removes the child node from the current node.
// If the child node is not a child of the current node, no action is taken.
//
// Parameters:
// - child: The child node to be removed.
func (n *NodeBase) RemoveChildNode(child Node) {
	n.children.Remove(child)
}

func (n *NodeBase) InsertChildrenAt(idx int, child Node) {
	if child == nil {
		panic("child is nil")
	}

	if idx > n.children.Len() {
		idx = n.children.Len()
	}

	previousIndex := n.children.IndexOf(child)

	if previousIndex != -1 {
		n.children.RemoveAt(previousIndex)

		if idx > previousIndex {
			idx--
		}
	}

	n.children.InsertAt(idx, child)
}

func (n *NodeBase) InsertChildBefore(anchor, node Node) {
	idx := n.children.IndexOf(anchor)

	if idx == -1 {
		return
	}

	n.InsertChildrenAt(idx, node)
}

func (n *NodeBase) InsertChildAfter(anchor, node Node) {
	idx := n.children.IndexOf(anchor)

	if idx == -1 {
		return
	}

	n.InsertChildrenAt(idx+1, node)
}

// ReplaceChildNode replaces an old child node with a new child node in the current node.
// If the old child node is not a child of the current node, no action is taken.
// The old child node is first removed from its parent node and detached from the graph.
// Then, the new child node is set as the replacement at the same index in the list of children nodes of the current node.
// The new child node is attached to the same graph as the parent node.
// Finally, any edges in the current node that reference the old child node as the destination node are updated to reference the new child node.
//
// Parameters:
// - old: The old child node to be replaced.
// - new: The new child node to replace the old child node.
func (n *NodeBase) ReplaceChildNode(old, new Node) {
	idx := n.children.IndexOf(old)

	if idx != -1 {
		n.children.Set(idx, new)
	}

	for it := n.edges.Iterator(); it.Next(); {
		kv := it.Item()
		e := kv.Value

		if e.To() == old {
			e = e.ReplaceTo(new)
		} else {
			continue
		}

		n.edges.Set(kv.Key, e)
	}
}
