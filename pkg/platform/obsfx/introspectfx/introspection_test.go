package introspectfx

import (
	"runtime/debug"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/npnsd/npnsd-kernel/pkg/kernel"

	obsfx "github.com/greenboxal/agibootstrap/pkg/platform/obsfx/collectionsfx"
)

type testType struct {
	foo string

	Foo string
	Bar int

	Names obsfx.MutableSlice[string]

	Kernel *kernel.Kernel
}

func TestIntrospect(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("PANIC: %v\nStack trace:\n%s", e, string(debug.Stack()))
			t.FailNow()
		}
	}()

	RegisterPropertyMethod[kernel.Kernel]("Identity")
	RegisterPropertyMethod[kernel.Kernel]("PrivateKey")
	RegisterPropertyMethod[kernel.Kernel]("ServiceProvider")
	RegisterPropertyMethod[kernel.Kernel]("ServiceManager")
	RegisterPropertyMethod[kernel.Kernel]("Options")
	RegisterPropertyMethod[kernel.Kernel]("TrustMasterKey")
	RegisterPropertyMethod[kernel.Kernel]("ReleaseLink")
	RegisterPropertyMethod[kernel.Kernel]("Release")
	RegisterPropertyMethod[kernel.Kernel]("SignedRelease")
	RegisterPropertyMethod[kernel.Kernel]("ReleaseInfo")
	RegisterPropertyMethod[kernel.Kernel]("RootFS")
	RegisterPropertyMethod[kernel.Kernel]("InitrdFS")

	v1 := &testType{
		foo: "foo",
		Foo: "bbq",
		Bar: 42,
	}

	v1.Names.Add("wtf")
	v1.Names.Add("bbq")

	typ := TypeOf(v1)

	if typ.Name() != "testType" {
		t.Errorf("unexpected type name: %s", typ.Name())
	}

	props := typ.Properties()

	sort.SliceStable(props, func(i, j int) bool {
		return props[i].Name() < props[j].Name()
	})

	barProp := typ.Property("Bar")
	fooProp := typ.Property("Foo")

	require.Contains(t, props, barProp)
	require.Contains(t, props, fooProp)

	require.Equal(t, barProp.Name(), "Bar")
	require.Equal(t, fooProp.Name(), "Foo")

	val := ValueOf(v1)

	require.Equal(t, props[0].GetValue(val).Go().Int(), int64(42))
	require.Equal(t, props[1].GetValue(val).Go().String(), "bbq")

	node := Introspect(v1)

	require.False(t, node.IsLeaf())
}
