package typesystem

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeName(t *testing.T) {
	testCase := func(v reflect.Value) func(t *testing.T) {
		return func(t *testing.T) {
			name := GetTypeName(v.Type())
			mangled := name.MangledName()
			unmangled, rest := ParseMangledName(mangled)

			fmt.Printf("%s\t=\t%s\n", name, mangled)

			require.Equal(t, name.Name, unmangled.Name)
			require.Equal(t, name.Package, unmangled.Package)
			require.EqualValues(t, name.InParameters, unmangled.InParameters)
			require.EqualValues(t, name.OutParameters, unmangled.OutParameters)
			require.Empty(t, rest)
		}
	}

	t.Run("fnArgsRet", testCase(reflect.ValueOf(func(a int, b any) (int, error) { panic("unreachable") })))
	t.Run("int", testCase(reflect.ValueOf(0)))
	t.Run("fn", testCase(reflect.ValueOf(func() {})))
	t.Run("fnArgs", testCase(reflect.ValueOf(func(a int, b any) {})))
}
