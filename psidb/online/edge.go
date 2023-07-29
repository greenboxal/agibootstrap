package online

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

type LiveEdge struct {
	psi.EdgeBase

	key    psi.EdgeKey
	from   *LiveNode
	to     *LiveNode
	dentry *graphfs.CacheEntry
	frozen *graphfs.SerializedEdge

	cookie int64
}

func (le *LiveEdge) From() psi.Node         { return le.from.node }
func (le *LiveEdge) Key() psi.EdgeReference { return le.key }

func (le *LiveEdge) ReplaceTo(node psi.Node) psi.Edge {
	to, err := le.from.g.Add(context.Background(), node)

	if err != nil {
		panic(err)
	}

	le.to = to

	return le
}

func (le *LiveEdge) To() psi.Node {
	if le.to == nil {
		to, err := le.ResolveTo(context.Background())

		if err != nil {
			panic(err)
		}

		return to
	}

	n, err := le.to.Get(context.Background())

	if err != nil {
		panic(err)
	}

	return n
}

func (le *LiveEdge) ResolveTo(ctx context.Context) (psi.Node, error) {
	le.resolve()

	return le.to.Get(ctx)
}

func NewLiveEdge(from *LiveNode, dentry *graphfs.CacheEntry) *LiveEdge {
	le := &LiveEdge{
		key:  dentry.Name().AsEdgeKey(),
		from: from,
	}

	le.Init(le)
	le.updateDentry(dentry)

	return le
}

func (le *LiveEdge) resolve() {
	if le.to != nil {
		return
	}

	le.to = le.from.g.nodeForDentry(le.dentry)
}

func (le *LiveEdge) invalidate() {
	le.to = nil
}

func (le *LiveEdge) update(se *graphfs.SerializedEdge) {
	le.frozen = se
}

func (le *LiveEdge) updateDentry(dentry *graphfs.CacheEntry) {
	if le.dentry == dentry {
		return
	}

	le.dentry = dentry
	le.frozen = nil
}

func (le *LiveEdge) Save(ctx context.Context, nh graphfs.NodeHandle) error {
	key := le.Key().GetKey()

	if le.frozen == nil {
		le.frozen = &graphfs.SerializedEdge{}
		le.frozen.Key = key
	}

	if key.Kind == psi.EdgeKindChild {
		le.frozen.Flags = graphfs.EdgeFlagRegular
	} else {
		le.frozen.Flags = graphfs.EdgeFlagLink
	}

	if le.to != nil {
		le.frozen.ToIndex = le.to.cachedIndex
		le.frozen.ToPath = le.to.cachedPath
	}

	return nh.SetEdge(ctx, le.frozen)
}
