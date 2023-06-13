//go:build selfwip
// +build selfwip

package psi

import "github.com/dave/dst"

type Node interface {
	ID() int64
	UUID() string
	Node() *NodeBase
	Parent() Node
	Children() []Node

	Ast() dst.Node

	IsContainer() bool
	IsLeaf() bool

	Comments() []string

	attachToGraph(g *Graph)
	detachFromGraph(g *Graph)
	setParent(parent Node)
	addChildNode(node Node)
	removeChildNode(node Node)
}
type Retriever struct {
}

func NewRetriever() *Retriever {
	return &Retriever{}
}

type Reference struct {
	Source Node
	Target Node

	DeclarationDistance int
	ReferenceDistance   int
}

type ReferenceIterator interface {
	Next() bool
	Reference() Reference
}

type referenceIteratorImpl struct {
	root Node

	current Reference
}

func (r *referenceIteratorImpl) Next() bool {
	stack := []Node{r.root}        // Initialize a stack with the root node
	visited := make(map[Node]bool) // Initialize a map to keep track of visited nodes

	for len(stack) > 0 {
		current := stack[len(stack)-1] // Get the top element from the stack
		stack = stack[:len(stack)-1]   // Remove the top element from the stack

		if !visited[current] {
			visited[current] = true // Mark the current node as visited

			// Perform the logic to process the current node,
			// such as checking for references or adding them to the result

			// Push the children of the current node onto the stack
			if current.IsContainer() {
				children := current.Children()
				for i := len(children) - 1; i >= 0; i-- {
					stack = append(stack, children[i])
				}
			}
		}
	}

	// TODO: Return true or false based on whether there are any references left to process
	// TODO: Update the `current` field with the next reference

	r.current = Reference{} // Set the current reference to an empty value
	return false            // There are no more references to process
}

func (r *referenceIteratorImpl) Reference() Reference {
	return r.current
}

func (r *Retriever) Retrieve(root Node) (ReferenceIterator, error) {
	return &referenceIteratorImpl{
		root: root,
	}, nil
}
