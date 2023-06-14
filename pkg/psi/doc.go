package psi

/*
Package psi provides a graph-based representation of a file directory system, with a focus on code files.

PSI:
The Project Structure Interface (PSI) is a graph network representation of a file directory system. Each file in the directory is represented as a node in the graph, and the code within each file is treated as a child node connected to its parent file node. The PSI allows for easy traversal and analysis of the file directory system.

UAST:
The Universal Abstract Syntax Tree (UAST) is an essential part of the PSI. It represents the code within each file as an Abstract Syntax Tree (AST). Each AST is associated with its parent file node in the PSI graph. The UAST captures the structure and relationships between code elements such as declarations, references, and implementations.

Node Interface:
The Node interface is a fundamental part of the PSI. Every node in the PSI graph implements this interface. It provides methods to access information about the node, such as its parent, children, AST, and comments. Additionally, the Node interface allows for adding and removing nodes from the graph.

Usage:
To use the PSI package, first import the "github.com/<username>/psi" module. Then, create a new PSI graph using the NewGraph() function. Add nodes representing files and their corresponding ASTs to the graph using the AddNode() method. Finally, use the provided methods of the Node interface to explore and analyze the file directory system.

Node Types:
The PSI package provides several node types that can be used to represent different elements of code files:

- The psi.Node interface represents a generic node in the PSI graph. It defines methods like Parent(), Children(), Ast(), etc. that can be used to access information about the node.
- The *psi.BaseNode type is a base implementation of the psi.Node interface. It provides common functionality for all PSI nodes.
- The *psi.Container type represents a container node in the PSI graph. It is used to represent code elements that can contain other code elements, like functions, classes, etc.
- The *psi.Leaf type represents a leaf node in the PSI graph. It is used to represent code elements that cannot contain other code elements, like variables, constants, etc.

Examples:
Here are some examples of how to use the PSI package:

// Create a new PSI graph
graph := psi.NewGraph()

// Create a file node
fileNode := psi.NewFileNode("example.go")

// Add the file node to the graph
graph.AddNode(fileNode)

// Create an AST node
astNode := psi.NewASTNode(fileNode)

// Add the AST node to the file node
fileNode.AddChildNode(astNode)

This is just a basic overview of the PSI package and its components. More detailed documentation can be found in the source code and accompanying documentation.
*/
func init() {}
