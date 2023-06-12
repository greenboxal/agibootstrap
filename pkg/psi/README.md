# The PSI and the UAST

## Overview

The goal of this project is to represent a file directory system, especially ones containing code files, as a graph network (G) in Go. Each file will be a node in the graph, and any Abstract Syntax Tree (AST) corresponding to the file's code will be considered a child of that file node. Edges in the graph will represent relationships between files and code declarations, references, and implementations.

We will define the Node interface, `psi.Node`, that every node N in the graph will implement. Each Node has a `Parent()` method that returns its parent node in the graph, corresponding to the `E_T_ParentDeclaration` edge type.

## Design

Here's the proposed design for our Node interface and related components:

```go
package psi

// Node is an interface for any node N in the graph G.
type Node interface {
    // Parent returns the parent node in the graph,
    // corresponding to the E_T_ParentDeclaration edge.
    Parent() Node
    // Additional methods may be added as needed to handle other edge types
}

// We can also define other node types such as FileNode, ASTNode, etc. as needed.

type FileNode struct {
    // ... other properties
}

func (f *FileNode) Parent() Node {
    // implement method to return parent of FileNode
}

type ASTNode struct {
    // ... other properties
}

func (a *ASTNode) Parent() Node {
    // implement method to return parent of ASTNode
}
```

## Usage

The graph G could be constructed by iterating over the file directory, creating a `FileNode` for each file, and `ASTNode` instances for each corresponding AST. The parent-child relationships would then be established according to the file directory hierarchy and AST structure.

## Metrics

We will also define various metrics such as `M_DeclarationDistance`, `M_ReferenceDistance`, `M_ImportanceDistance`, and `M_ImportanceScore` to evaluate relationships between nodes in the graph. Implementations of these metrics would depend on the graph's structure and specific edge definitions, and could be added to the `Node` interface or defined as standalone functions.

## Future Work

Future work might involve handling different edge types besides `E_T_ParentDeclaration`, adding more node types, and improving the performance of metric calculations. It may also include extending the graph representation to handle more complex structures like multi-root graphs and cyclical dependencies.

## Conclusion

This design document provides a basic structure for implementing a graph representation of a file directory system in Go. The actual implementation will likely involve additional details and considerations, but this serves as a solid starting point.

## Glossary

* PSI: Project Structure Interface
* UAST: Universal Abstract Syntax Tree
