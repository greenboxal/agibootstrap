package indexing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Entry struct {
	psi.NodeBase

	LocalIndex  int64 `json:"L"`
	RemoteIndex int64 `json:"R"`
	OriginIndex int64 `json:"O"`
}

var EntryEdge = psi.DefineEdgeType[*Entry]("psidb.indexing.entry")
var EntryOriginEdge = psi.DefineEdgeType[psi.Node]("psidb.indexing.entry.origin")
var EntryType = psi.DefineNodeType[*Entry]()

func (e *Entry) PsiNodeName() string { return strconv.FormatInt(e.LocalIndex, 10) }

func NewEntry(localKey int64) *Entry {
	e := &Entry{}
	e.LocalIndex = localKey
	e.RemoteIndex = -1
	e.Init(e, psi.WithNodeType(EntryType))

	return e
}

type Index struct {
	psi.NodeBase

	Name string `json:"name"`

	IndexManager *Manager   `inject:"" json:"-"`
	Index        *LiveIndex `json:"-"`
}

func (idx *Index) SearchEntries(ctx context.Context, query []float32, k int) ([]*Entry, []float32, error) {
	li, err := idx.GetLiveIndex(ctx)

	if err != nil {
		return nil, nil, err
	}

	indexes, distances, err := li.Search(query, k)

	entries := make([]*Entry, len(indexes))

	for i, v := range indexes {
		entry := idx.ResolveChild(ctx, psi.PathElement{Name: strconv.FormatInt(v, 10)})

		if entry == nil {
			return nil, nil, fmt.Errorf("entry not found")
		}

		entries[i] = entry.(*Entry)
	}

	return entries, distances, nil
}

func (idx *Index) IndexEntry(ctx context.Context, item *Embedding) error {
	li, err := idx.GetLiveIndex(ctx)

	if err != nil {
		return err
	}

	for i, v := range item.V {
		edgeKey := EntryEdge.NamedIndexed(idx.Name, int64(i))
		entry := psi.ResolveEdge(idx, edgeKey)

		if entry == nil {
			entry = NewEntry(time.Now().UnixNano())
			entry.SetParent(idx)
			entry.SetEdge(EntryOriginEdge.Singleton(), item)

			item.SetEdge(edgeKey, entry)
		}

		entry.OriginIndex = int64(i)

		if entry.RemoteIndex == -1 {
			if err := li.Remove(entry.LocalIndex); err != nil {
				return err
			}

			entry.RemoteIndex = -1
		}

		if err := li.Add(entry.LocalIndex, v); err != nil {
			return err
		}

		if err := entry.Update(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (idx *Index) GetLiveIndex(ctx context.Context) (*LiveIndex, error) {
	if idx.Index == nil {
		li, err := idx.IndexManager.OpenIndex(ctx, idx.Name, 1536)

		if err != nil {
			return nil, err
		}

		idx.Index = li
	}

	return idx.Index, nil
}
