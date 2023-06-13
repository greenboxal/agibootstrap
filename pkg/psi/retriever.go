//go:build selfwip
// +build selfwip

package psi

// G is a graph network, analogous to a file directory system containing files. For each file that contains code, the Abstract Syntax Tree (AST) for that code is treated as a child node connected to the parent node that represents the file.
//
// There are different types of subgraphs, G_N, in G, each type corresponding to a unique type of edge, E_T. Here are the specifics:
//
// 1. Subgraph G_N_Declaration: This corresponds to edge type E_T_ParentDeclaration. This relationship essentially maps each code declaration in a file (a node) to the file it belongs to (its parent node). Conversely, there is a one-to-many mapping known as E_T_ChildrenDeclarations, which relates a file (a node) to all the code declarations it contains (its child nodes).
//
// 2. Subgraph G_N_References: This subgraph denotes the connections between various nodes in an AST based on references. References can be of different types like type references (locations in code where a certain data type is used such as variable declarations or function parameters) and function references (points in the code where a function is called and where it is declared).
//
// 3. Subgraph G_N_Implementations: This subgraph represents the relations between nodes that denote a virtual function and all its respective implementations, or between an abstract/virtual interface and all its implementations.
//
// The metric M represents the distance between two nodes in a graph. In this context, it could represent the relationship or 'distance' between different components of a codebase, such as between a function declaration and its call site, between an interface and its implementations, or between a type and its usage sites.
//
// 1. M_DeclarationDistance: This metric measures the number of hops between two nodes N_Ancestor and N_Child within the subgraph G_N_Declaration. A 'hop' represents a single step from one node to a directly adjacent node. This metric essentially quantifies the degree of separation between a code declaration and its parent file in the directory hierarchy.
//
// 2. M_ReferenceDistance: This metric quantifies the number of hops between two nodes N_ReferenceRoot and N_Referenced within the subgraph G_N_References. The N_ReferenceRoot represents the origin of the reference, while N_Referenced is the target of the reference. This metric measures the degree of separation between two pieces of code that are linked through type or function references.
//
// 3. M_ImportanceDistance: This is a combined metric that considers both M_DeclarationDistance and M_ReferenceDistance, representing them in a two-dimensional vector space. Each point in this 2D space corresponds to a code component, with its coordinates representing the declaration distance and reference distance of that component respectively.
//
// 4. M_ImportanceScore: This metric is defined as the inverse of the norm of the M_ImportanceDistance vector subtracted from one (1 - norm(M_ImportanceDistance)). The norm here refers to the vector's magnitude, which can be computed as the square root of the sum of squares of its components. The M_ImportanceScore is a measure of the relative significance or 'importance' of a code component in the codebase, taking into consideration both its declaration hierarchy and its references within the codebase. A higher score indicates a higher level of 'importance' or centrality in the codebase.

// Reference:
//type Node interface {
//	ID() int64
//	UUID() string
//	Node() *NodeBase
//	Parent() Node
//	Children() []Node
//
//	Ast() dst.Node
//
//	IsContainer() bool
//	IsLeaf() bool
//
//	Comments() []string
//
//	attachToGraph(g *Graph)
//	detachFromGraph(g *Graph)
//	setParent(parent Node)
//	addChildNode(node Node)
//	removeChildNode(node Node)
//}

// Retriever is an interface that can be used to retrieve nodes from a graph
type Retriever struct {
}

// NewRetriever creates a new instance of the Retriever
//
//	func NewRetriever() *Retriever {
//		return &Retriever{}
//	}
func NewRetriever() *Retriever {
	return &Retriever{}
}

// ReferenceIteratorImpl is an implementation of the ReferenceIterator interface
type ReferenceIteratorImpl struct {
	root    Node
	current Reference
}

// Reference represents a reference between two nodes
type Reference struct {
	// Referer is the node that is referencing the
	Referer Node
	// Referee is the node that is being referenced
	Referee Node

	// DeclarationDistance is the distance between the declaration of the referee and the reference
	DeclarationDistance int
	// ReferenceDistance is the distance between the reference and the declaration of the referee
	ReferenceDistance int
}

// ReferenceIterator is an iterator that can be used to iterate over references
type ReferenceIterator interface {
	// Next returns true if there are any references left to process
	Next() bool
	// Reference returns the current reference
	Reference() Reference
}

// ReferenceIteratorImpl is an implementation of the ReferenceIterator interface
type referenceIteratorImpl struct {
	root Node

	current Reference
}

// TODO: Implement this method properly.
func (r *referenceIteratorImpl) Next() bool {
	// Implementation to be added
	// Code goes here
	// Updated implementation
	if len(r.root.Children()) == 0 {
		return false
	}
	// More code
	return true
}

// Reference returns the current reference
func (r *referenceIteratorImpl) Reference() Reference {
	return r.current
}

// Retrieve returns a reference iterator that can be used to iterate over all references in the graph
// starting from the given root node.
func (r *Retriever) Retrieve(root Node) (ReferenceIterator, error) {
	return &referenceIteratorImpl{
		root: root,
	}, nil
}
