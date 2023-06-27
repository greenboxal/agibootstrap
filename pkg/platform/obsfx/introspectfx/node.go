package introspectfx

import (
	"reflect"

	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx/collectionsfx"
)

type Path []string

type Node interface {
	Value

	IsLeaf() bool
	Children() collectionsfx.ObservableList[Node]
}

type valueNode struct {
	Value

	children collectionsfx.MutableSlice[Node]
	props    collectionsfx.MutableMap[string, Node]

	hasInitializedProps bool
}

func newValueNode(value Value) Node {
	if n, ok := value.(Node); ok {
		return n
	}

	n := &valueNode{
		Value: value,
	}

	collectionsfx.BindListFromMap(&n.children, &n.props, func(k string, v Node) Node {
		return v
	})

	return n
}

func (n *valueNode) IsLeaf() bool {
	return false
}

func (n *valueNode) Children() collectionsfx.ObservableList[Node] {
	if !n.hasInitializedProps {
		for _, prop := range n.Type().Properties() {
			var child Node

			if prop.IsList() {
				child = newListNode(prop.GetValue(n))
			} else {
				child = newValueNode(prop.GetValue(n))
			}

			n.props.Set(prop.Name(), child)
		}

		n.hasInitializedProps = true
	}

	return &n.children
}

type listNode struct {
	Value

	children collectionsfx.MutableSlice[Node]

	hasInitializedProps bool
}

func newListNode(v Value) Node {
	n := &listNode{
		Value: v,
	}

	return n
}

func (n *listNode) IsLeaf() bool {
	return false
}

func (n *listNode) Children() collectionsfx.ObservableList[Node] {
	if !n.hasInitializedProps {
		src := n.Go()

		if src.Kind() != reflect.Ptr {
			src = src.Addr()
		}

		obs := src.Interface().(collectionsfx.BasicObservableList)

		collectionsfx.BindListAny(&n.children, obs, func(v any) Node {
			return newValueNode(ValueOf(v))
		})

		n.hasInitializedProps = true
	}

	return &n.children
}
