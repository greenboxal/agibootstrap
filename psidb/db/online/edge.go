package online

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	graphfs "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type LiveEdge struct {
	psi.EdgeBase

	key    psi.EdgeKey
	from   *LiveNode
	to     *LiveNode
	dentry *graphfs.CacheEntry
	frozen *graphfs.SerializedEdge

	dirty bool
}

func (le *LiveEdge) From() psi.Node         { return le.from.node }
func (le *LiveEdge) Key() psi.EdgeReference { return le.key }

func (le *LiveEdge) ReplaceTo(node psi.Node) psi.Edge {
	to := le.from.g.nodeForNode(node)

	if le.to == to {
		return le
	}

	le.to = to
	le.dirty = true

	return le
}

func (le *LiveEdge) To() psi.Node {
	to, err := le.ResolveTo(context.Background())

	if err != nil {
		panic(err)
	}

	return to
}

func (le *LiveEdge) ResolveTo(ctx context.Context) (psi.Node, error) {
	le.resolve()

	if le.to == nil {
		return nil, psi.ErrNodeNotFound
	}

	return le.to.Get(ctx)
}

func NewLiveEdge(from *LiveNode, key psi.EdgeKey) *LiveEdge {
	le := &LiveEdge{
		key:  key,
		from: from,
	}

	le.Init(le)

	return le
}

func (le *LiveEdge) resolve() {
	if le.to != nil {
		return
	}

	if le.dentry.IsNegative() {
		return
	}

	le.to = le.from.g.nodeForDentry(le.dentry)
}

func (le *LiveEdge) invalidate() {
	le.to = nil
}

func (le *LiveEdge) update(se *graphfs.SerializedEdge) {
	le.frozen = se
	le.dirty = false
	le.to = nil
}

func (le *LiveEdge) updateDentry(dentry *graphfs.CacheEntry) {
	if le.dentry == dentry {
		return
	}

	le.dentry = dentry
	le.frozen = nil
}

func (le *LiveEdge) Save(ctx context.Context, nh graphfs.NodeHandle) error {
	if !le.dirty {
		return nil
	}

	key := le.Key().GetKey()

	if le.frozen == nil {
		le.frozen = &graphfs.SerializedEdge{}
		le.frozen.Key = key
	}

	if le.key.Kind == "" || le.key.Kind == psi.EdgeKindChild {
		le.frozen.Flags &= ^graphfs.EdgeFlagModes
		le.frozen.Flags |= graphfs.EdgeFlagRegular
	} else {
		le.frozen.Flags &= ^graphfs.EdgeFlagModes
		le.frozen.Flags |= graphfs.EdgeFlagLink
	}

	if le.to != nil {
		le.frozen.ToIndex = le.to.cachedIndex
		le.frozen.ToPath = le.to.path
	}

	if le.frozen.Flags&graphfs.EdgeFlagRemoved == 0 {
		if err := nh.SetEdge(ctx, le.frozen); err != nil {
			return err
		}
	} else {
		if err := nh.RemoveEdge(ctx, key); err != nil {
			return err
		}
	}

	le.dirty = false

	return nil
}
