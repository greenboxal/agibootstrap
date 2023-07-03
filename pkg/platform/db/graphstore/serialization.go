package graphstore

import (
	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type nodeWrapper struct {
	Node psi.Node `json:"node"`
}

type edgeWrapper struct {
	Edge psi.Edge `json:"edge"`
}

var nodeWrapperType = typesystem.TypeOf(nodeWrapper{})
var edgeWrapperType = typesystem.TypeOf(edgeWrapper{})
var frozenNodeType = typesystem.TypeOf(&psi.FrozenNode{})
var frozenEdgeType = typesystem.TypeOf(&psi.FrozenEdge{})

var defaultLinkPrototype = cidlink.LinkPrototype{
	Prefix: cid.Prefix{
		Codec:    cid.DagJSON,
		MhLength: -1,
		MhType:   multihash.SHA2_256,
		Version:  1,
	},
}
