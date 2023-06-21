package rendering

import (
	"io"

	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

/*
# Revised Design Document: Rendering Operation for Tree T with Lazy Leaf Nodes and Token Limit

## 1. Introduction
This document proposes a design for a tree structure (referred to as "Tree T") where each parent node can have multiple ordered children nodes and each leaf node can have an associated string value, which is generated lazily at the time of tree traversal during the rendering operation.

In addition, each Node in Tree T has a weight (W) ranging from 0 to 1, acting as a normalized weight, and a priority (P) that can take any integer number. There's also a Token Limit, which determines the maximum length of the output string during the rendering operation.

## 2. Design Requirements

### 2.1 Node Structure
Each Node in Tree T will have the following properties:

1. **Value Generator**: For leaf nodes, this is a function or a promise that generates the node's value when called/executed. For non-leaf nodes, it could be undefined or have a specific type depending on the implementation.

2. **Children**: An ordered list of child nodes.

3. **Weight (W)**: A float between 0 and 1 inclusive. This represents the relative importance of the node in relation to its siblings.

4. **Priority (P)**: An integer representing the priority of the node. This can be used to determine the order in which nodes are processed during the Rendering Operation.

### 2.2 Tree Structure
Tree T should support the following operations:

1. **Add Node**: Add a new Node to Tree T with a given Value Generator, Weight, and Priority. If adding a child node, the parent node should be specified.

2. **Remove Node**: Remove a Node from Tree T. If the node has children, those nodes are also removed.

3. **Get Node**: Retrieve a Node from Tree T based on a given identifier.

4. **Update Node**: Update the Value Generator, Weight, or Priority of a Node.

5. **Rendering Operation**: Traverse the tree and write all leaf nodes to a single string buffer, with a token limit that caps the maximum length of the final string.

## 3. Rendering Operation Design

### 3.1 Overview
The Rendering Operation is a traversal of Tree T that writes all leaf nodes to a single string buffer. Leaf nodes are rendered first, with parent nodes appending content around and between their children. The operation takes into account a Token Limit that restricts the length of the output string.

### 3.2 Detailed Procedure
The Rendering Operation works as follows:

1. **Start at the Root Node**: The operation begins at the root node of Tree T.

2. **Initialize Token Counter**: The counter is set to the predefined Token Limit at the start.

3. **Node Prioritization**: In the event of multiple children at a parent node, children nodes are sorted based on their priority (P). In case of a tie in priority, the weight (W) is considered. Nodes with higher priority are processed first.

4. **Leaf Node Processing**: If the current node is a leaf node, its value generator function is called. If the length of the generated string is less than or equal to the remaining token count, it is appended to the string buffer and the token count is decreased by the length of the string. If the length exceeds the remaining token count, the operation halts or the string is truncated, depending on the specific implementation.

5. **Non-leaf Node Processing**: If the current node is a non-leaf node, the Rendering Operation recursively processes its children, continually checking and updating the remaining token count. Content may be appended to the string buffer before, after, or between the processing of child nodes, as
*/

const DoNotPrune = -1

type NodeState struct {
	Priority       int
	InitialWeight  float32
	CurrentWeight  float32
	ChildrenWeight float32
	TokenCount     int
	Buffer         *TokenBuffer
	Node           psi.Node
}

func NewNodeState(r *PruningRenderer, node psi.Node) *NodeState {
	return &NodeState{
		Node:          node,
		InitialWeight: 1.0,
		Buffer:        NewTokenBuffer(r.Tokenizer, 0),
	}
}

func (ns *NodeState) WriteTo(writer io.Writer) (total int64, err error) {
	return ns.Buffer.WriteTo(writer)
}

func (ns *NodeState) Update(renderer *PruningRenderer) error {
	ns.Reset(renderer.Tokenizer)

	_, err := renderer.Write(ns.Buffer, ns.Node)

	if err != nil {
		return err
	}

	ns.TokenCount = ns.Buffer.TokenCount()
	ns.InitialWeight = renderer.Weight(ns, ns.Node)

	return nil
}

func (ns *NodeState) Reset(tokenizer tokenizers.BasicTokenizer) {
	ns.Buffer = NewTokenBuffer(tokenizer, 0)
	ns.Buffer.state = ns
	ns.TokenCount = 0
}

type PruningRenderer struct {
	nodeStates map[string]*NodeState

	Tokenizer tokenizers.BasicTokenizer
	Weight    func(state *NodeState, node psi.Node) float32
	Write     func(w *TokenBuffer, node psi.Node) (int, error)
}

func (r *PruningRenderer) Render(node psi.Node, writer io.Writer) (total int64, err error) {
	if r.nodeStates == nil {
		r.nodeStates = make(map[string]*NodeState)
	}

	totalWeight := float32(0.0)

	err = psi.Walk(node, func(cursor psi.Cursor, entering bool) error {
		n := cursor.Node()
		s := r.getState(n)

		if entering {
			if n.IsLeaf() {
				if err := s.Update(r); err != nil {
					return err
				}

				s.CurrentWeight = s.InitialWeight

				totalWeight += s.CurrentWeight

				if p := n.Parent(); p != nil {
					ps := r.getState(n)

					if ps != nil {
						ps.ChildrenWeight += s.CurrentWeight
					}
				}
			} else {
				s.ChildrenWeight = 0
			}

		} else if n.IsContainer() {
			for _, c := range n.Children() {
				cs := r.getState(c)

				if cs != nil {
					cs.CurrentWeight /= s.ChildrenWeight
				}
			}

			if err := s.Update(r); err != nil {
				return err
			}

			s.CurrentWeight = s.InitialWeight * s.ChildrenWeight

			totalWeight += s.CurrentWeight
		}

		return nil
	})

	if err != nil {
		return
	}

	targetState := r.getState(node)

	if err := targetState.Update(r); err != nil {
		return total, err
	}

	n, err := writer.Write(targetState.Buffer.Bytes())

	if err != nil {
		return total, err
	}

	total += int64(n)

	return
}

func (r *PruningRenderer) getState(n psi.Node) *NodeState {
	s, ok := r.nodeStates[n.UUID()]

	if !ok {
		s = &NodeState{}
		s.Node = n
		s.Reset(r.Tokenizer)
		r.nodeStates[n.UUID()] = s
	}

	return s
}
