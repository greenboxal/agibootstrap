package thoughtstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchSimple(t *testing.T) {
	branch := NewBranchFromSlice(nil, RootPointer())

	t1 := NewThought()
	t2 := NewThought()
	t3 := NewThought()

	branchStream := branch.Mutate()
	branchStream.Append(t1)
	branchStream.Append(t2)
	branchStream.Append(t3)

	require.Equal(t, 3, branch.Len())

}

func TestThoughtStream(t *testing.T) {
	// Initialize Resolver
	r := newTestResolver()

	// Create a base Branch
	basePtr := Pointer{
		Parent:    cid.Cid{},
		Previous:  cid.Cid{},
		Timestamp: time.Now(),
		Level:     1,
		Clock:     1,
	}
	rootThought := NewThought()
	rootThought.Pointer = basePtr

	rootBranch := NewBranchFromSlice(nil, basePtr, rootThought)

	// Test base branch properties
	assert.Equal(t, basePtr, rootBranch.BasePointer())
	assert.Nil(t, rootBranch.Base())

	// Add a Thought to base branch
	baseThought := NewThought()
	baseBranchStream := rootBranch.Mutate()
	baseBranchStream.Append(baseThought)

	// Create a new Branch from base Branch
	newThought := NewThought()

	newBranchStream := baseBranchStream.Fork()
	newBranchStream.Append(newThought)

	newBranch := newBranchStream.AsBranch()

	r.AddBranch(rootBranch)
	r.AddBranch(newBranch)

	// Merge baseBranchStream into newBranchStream
	err := baseBranchStream.Merge(context.Background(), r, HierarchicalTimeMergeStrategy(r), newBranch)

	assert.Nil(t, err)
}

type testResolver struct {
	branches map[cid.Cid]Branch
	thoughts map[cid.Cid]*Thought
}

func newTestResolver() *testResolver {
	return &testResolver{
		branches: make(map[cid.Cid]Branch),
		thoughts: make(map[cid.Cid]*Thought),
	}
}

func (r *testResolver) AddBranch(b Branch) {
	r.branches[b.BasePointer().Address()] = b

	for it := b.Stream(); it.Next(); {
		r.AddThought(it.Value())
	}
}

func (r *testResolver) AddThought(thought *Thought) {
	r.thoughts[thought.Pointer.Address()] = thought
}

func (r *testResolver) ResolveThought(ctx context.Context, id cid.Cid) (*Thought, error) {
	t := r.thoughts[id]

	if t == nil {
		return nil, fmt.Errorf("branch not found: %s", id)
	}

	return t, nil
}

func (r *testResolver) ResolveBranch(ctx context.Context, id cid.Cid) (Branch, error) {
	b := r.branches[id]

	if b == nil {
		return nil, fmt.Errorf("branch not found: %s", id)
	}

	return b, nil
}
