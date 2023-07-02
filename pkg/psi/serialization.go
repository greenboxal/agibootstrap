package psi

import (
	"encoding/json"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

type PhysAddr = cid.Cid

type RawPointer struct {
	Depth   int      `json:"Depth"`
	Index   int      `json:"Index"`
	Parent  PhysAddr `json:"Parent"`
	Thought PhysAddr `json:"Thought"`
}

type RawAttribute struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type RawEdge struct {
	Key EdgeKey `json:"key"`
	To  Path    `json:"to"`
}

type RawNodeEntry struct {
	UUID     string   `json:"UUID,omitempty"`
	PhysAddr PhysAddr `json:"PhysAddr,omitempty"`
	Node     Node     `json:"Value,omitempty"`
}

type RawNode struct {
	ID         int64          `json:"ID"`
	UUID       string         `json:"UUID"`
	Attributes []RawAttribute `json:"Attributes"`
	Edges      []RawEdge      `json:"Edges"`
	Children   []RawNodeEntry `json:"Children"`
}

type Freezer struct {
	Cas   map[cid.Cid][]byte  `json:"Cas"`
	Cache map[cid.Cid]RawNode `json:"-"`
	IdMap map[NodeID]cid.Cid  `json:"IdMap"`
}

func (f *Freezer) GetByID(id NodeID) (RawNode, bool) {
	addr, ok := f.IdMap[id]

	if !ok {
		return RawNode{}, false
	}

	return f.Get(addr)
}

func (f *Freezer) Get(id cid.Cid) (RawNode, bool) {
	node, ok := f.Cache[id]

	if !ok {
		data, ok := f.Cas[id]

		if !ok {
			return RawNode{}, false
		}

		err := json.Unmarshal(data, &node)

		if err != nil {
			return RawNode{}, false
		}

		f.Cache[id] = node
	}

	return node, ok
}

func (f *Freezer) Add(n Node) (entry RawNodeEntry) {
	if _, ok := f.IdMap[n.CanonicalPath().String()]; ok {
		return
	}

	frozen := RawNode{}
	frozen.ID = n.ID()
	frozen.UUID = n.CanonicalPath().String()
	frozen.Children = make([]RawNodeEntry, len(n.Children()))
	frozen.Edges = make([]RawEdge, 0)
	frozen.Attributes = make([]RawAttribute, 0)

	for i, child := range n.Children() {
		childEntry := f.Add(child)

		frozen.Children[i] = childEntry
	}

	for it := n.Edges(); it.Next(); {
		frozen.Edges = append(frozen.Edges, RawEdge{
			Key: it.Edge().Key().GetKey(),
			To:  it.Edge().To().CanonicalPath(),
		})
	}

	data, err := json.Marshal(frozen)

	if err != nil {
		panic(err)
	}

	mh, err := multihash.Sum(data, multihash.SHA2_256, -1)
	addr := cid.NewCidV1(cid.Raw, mh)

	f.Cas[addr] = data
	f.Cache[addr] = frozen
	f.IdMap[n.CanonicalPath().String()] = addr

	return
}
