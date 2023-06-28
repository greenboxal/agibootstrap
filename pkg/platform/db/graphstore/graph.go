package graphstore

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Store interface {
	// GetNodesByID returns the nodes with the given ids.
	GetNodesByID(ids []psi.NodeID) ([]psi.Node, error)

	// GetEdgesByID returns the edges with the given ids.
	GetEdgesByID(nodeId psi.NodeID, ids []psi.EdgeID) ([]psi.Edge, error)

	// RemoveEdges removes the edges with the given ids.
	RemoveEdges(ids []psi.EdgeID) error
}
