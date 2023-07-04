package thoughtdb

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func TestBranch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootNode := &psi.NodeBase{}
	rootNode.Init(rootNode)

	g := graphstore.NewIndexedGraph(datastore.NewMapDatastore(), rootNode)
	r := NewRepo(g)
	b := r.CreateBranch()

	t1 := NewThought()
	t2 := NewThought()
	t3 := NewThought()

	err := b.Commit(ctx, t1)
	require.NoError(t, err)

	err = b.Commit(ctx, t2)
	require.NoError(t, err)

	err = b.Commit(ctx, t3)
	require.NoError(t, err)

	cursor := b.Cursor()
	cursor.PushParents()

	//require.Equal(t, t3, cursor.Thought())

	parentsCursor := b.Cursor()
	parentIterator := parentsCursor.IterateParents()
	parents := iterators.ToSlice(parentIterator)

	require.Equal(t, []Pointer{t2.Pointer, t1.Pointer}, parents)
}
