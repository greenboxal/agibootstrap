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

Please note that this is a basic overview and introduction to the PSI package. More detailed documentation can be found in the source code and accompanying documentation.
*/

func init() {}
