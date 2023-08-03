package psidsadapter

import (
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

var dsKeyBitmap = psids.Key[BitmapSnapshot]("_graph/bitmap")
var dsKeyNodeData = psids.KeyTemplate[*graphfs.SerializedNode]("node-state/%d")
var dsKeyNodeEdge = psids.KeyTemplate[*graphfs.SerializedEdge]("node-edge/%d!/%s")
var dsKeyEdgePrefix = psids.KeyTemplate[*graphfs.SerializedEdge]("node-edge/%d!")

var defaultLinkPrototype = cidlink.LinkPrototype{
	Prefix: cid.Prefix{
		Codec:    cid.DagJSON,
		MhLength: -1,
		MhType:   multihash.SHA2_256,
		Version:  1,
	},
}
