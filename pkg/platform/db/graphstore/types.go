package graphstore

import (
	"reflect"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
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

var lastFenceKey = datastore.NewKey("_graphstore/_lastFence")
var frozenNodeType = typesystem.TypeFrom(reflect.TypeOf((*psi.FrozenNode)(nil)).Elem())
var frozenEdgeType = typesystem.TypeFrom(reflect.TypeOf((*psi.FrozenEdge)(nil)).Elem())

var defaultLinkPrototype = cidlink.LinkPrototype{
	Prefix: cid.Prefix{
		Codec:    cid.DagJSON,
		MhLength: -1,
		MhType:   multihash.SHA2_256,
		Version:  1,
	},
}

var NoDataCid cid.Cid

func init() {
	mh, err := multihash.Sum(nil, multihash.SHA2_256, -1)

	if err != nil {
		panic(err)
	}

	NoDataCid = cid.NewCidV1(cid.Raw, mh)
}
