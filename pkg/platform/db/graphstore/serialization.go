package graphstore

import (
	"encoding/json"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FrozenGraph struct {
	Nodes []*FrozenNode
	Edges []*FrozenEdge
}

type FrozenNode struct {
	Cid cidlink.Link `json:"cid,omitempty"`

	Index   int64 `json:"index"`
	Version int64 `json:"version"`

	UUID psi.NodeID   `json:"uuid"`
	Type psi.NodeType `json:"type"`

	Attributes map[string]interface{} `json:"attr,omitempty"`
}

type FrozenEdge struct {
	Cid cidlink.Link `json:"cid,omitempty"`

	Key psi.EdgeKey `json:"key"`

	From psi.NodeID `json:"from"`
	To   psi.NodeID `json:"to"`

	Attributes map[string]interface{} `json:"attr,omitempty"`
}

type nodeWrapper struct {
	Node psi.Node `json:"node"`
}

type edgeWrapper struct {
	Edge psi.Edge `json:"edge"`
}

var nodeWrapperType = typesystem.TypeOf(nodeWrapper{})
var edgeWrapperType = typesystem.TypeOf(edgeWrapper{})
var frozenNodeType = typesystem.TypeOf(&FrozenNode{})
var frozenEdgeType = typesystem.TypeOf(&FrozenEdge{})

func SerializeNode(n psi.Node) ([]byte, cid.Cid, error) {
	wrapped := typesystem.Wrap(nodeWrapper{Node: n})

	data, err := ipld.Encode(wrapped, dagjson.Encode)

	if err != nil {
		return nil, cid.Undef, err
	}

	mh, err := multihash.Sum(data, multihash.SHA2_256, -1)

	if err != nil {
		return nil, cid.Undef, err
	}

	id := cid.NewCidV1(cid.Raw, mh)

	return data, id, nil
}

func DeserializeNode(uuid psi.NodeID, data []byte) (psi.Node, error) {
	wrapped, err := ipld.DecodeUsingPrototype(data, dagjson.Decode, nodeWrapperType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	n := typesystem.Unwrap(wrapped).(nodeWrapper).Node

	if n != nil {
		n.PsiNodeBase().Init(n, uuid)
	}

	return n, nil
}

func FreezeGraph(g psi.Graph) (*FrozenGraph, error) {
	var fg FrozenGraph

	for it := g.Nodes(); it.Next(); {
		n := it.Node()

		_, contentId, err := SerializeNode(n)

		if err != nil {
			return nil, err
		}

		fn := &FrozenNode{
			Cid: cidlink.Link{Cid: contentId},

			Index:      n.ID(),
			UUID:       n.UUID(),
			Type:       n.PsiNodeType(),
			Attributes: n.Attributes(),
		}

		fg.Nodes = append(fg.Nodes, fn)

		childrenIndex := int64(0)

		for it := n.ChildrenIterator(); it.Next(); childrenIndex++ {
			cn := it.Node()

			fe := &FrozenEdge{
				Key: psi.EdgeKey{
					Kind:  psi.EdgeKindChild,
					Index: childrenIndex,
				},

				From: n.UUID(),
				To:   cn.UUID(),
			}

			fg.Edges = append(fg.Edges, fe)
		}

		for edges := n.Edges(); edges.Next(); {
			e := edges.Edge()

			fe := &FrozenEdge{
				Key:  e.Key().GetKey(),
				From: e.From().UUID(),
				To:   e.To().UUID(),
			}

			fg.Edges = append(fg.Edges, fe)
		}
	}

	return &fg, nil
}

func SerializeGraph(g psi.Graph) ([]byte, error) {
	fg, err := FreezeGraph(g)

	if err != nil {
		return nil, err
	}

	return json.Marshal(fg)
}

func DeserializeGraph(data []byte) (*FrozenGraph, error) {
	var fg FrozenGraph

	err := json.Unmarshal(data, &fg)
	if err != nil {
		return nil, err
	}

	return &fg, nil
}
