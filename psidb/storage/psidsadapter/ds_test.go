package psidsadapter

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

func TestNewDataStoreSuperBlock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ds := datastore.NewMapDatastore()
	sb := NewDataStoreSuperBlock(ds, "uuid")

	vfs := graphfs.NewVirtualGraph(func(ctx context.Context, uuid string) (graphfs.SuperBlock, error) {
		if uuid == "uuid" {
			return sb, nil
		}

		return nil, nil
	})

	rootPath := psi.PathFromElements("uuid", false)

	sb2, err := vfs.GetSuperBlock(ctx, rootPath.Root())

	require.NoError(t, err)
	require.Equal(t, sb, sb2)

	sb3, err := vfs.GetSuperBlock(ctx, "uuid2")

	require.NoError(t, err)
	require.Nil(t, sb3)

	rootCe, err := vfs.Resolve(ctx, rootPath)

	require.NoError(t, err)
	require.NotNil(t, rootCe)

	childPath := rootPath.Join(psi.MustParsePath("foo"))

	ce, err := vfs.Resolve(ctx, childPath)

	require.NoError(t, err)
	require.NotNil(t, ce)
	require.True(t, ce.IsNegative())

	nh, err := vfs.Open(ctx, childPath, graphfs.WithOpenNodeFlag(graphfs.OpenNodeFlagsCreate))

	require.NoError(t, err)
	require.NotNil(t, nh)

	fe := &psi.FrozenNode{}
	_, err = nh.Write(ctx, fe)

	require.NoError(t, err)

	fe2, err := nh.Read(ctx)

	require.NoError(t, err)
	require.Equal(t, *fe, *fe2)

	err = nh.Close()

	require.NoError(t, err)
}
