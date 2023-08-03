package online

import (
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
)

var linkPrototype = cidlink.LinkPrototype{
	Prefix: cid.Prefix{
		Version:  1,
		Codec:    cid.DagCBOR,
		MhType:   multihash.MD5,
		MhLength: -1,
	},
}
