package graphstore

import (
	"reflect"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type edgeWrapper struct {
	Edge psi.Edge `json:"edge"`
}

var dsKeyLastFence = psids.Key[uint64]("_graphstore/_lastFence")
var dsKeyBitmap = psids.Key[BitmapSnapshot]("_graphstore/_bitmap")
var dsKeyRootUuid = psids.Key[string]("_graphstore/_rootUuid")
var dsKeyRootPath = psids.Key[psi.Path]("_graphstore/_rootPath")
var dsKeyRootSnapshot = psids.Key[cidlink.Link]("_graphstore/_rootSnap")

var dsKeyNodeHead = psids.KeyTemplate[cidlink.Link]("refs/heads/%s")
var dsKeyEdge = psids.KeyTemplate[cidlink.Link]("refs/edges/%s!/%s")
var dsKeyEdgePrefix = psids.KeyTemplate[cidlink.Link]("refs/edges/%s!")

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
