package vm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const srcFileName = "/Users/jonathanlima/mark/agibootstrap/staging/sdk/psidb-ts/dist/main.js"
const srcFilename2 = "/Users/jonathanlima/mark/agibootstrap/staging/sdk/psidb-ts/dist/vendors-node_modules_react_index_js.js"

func srcFileContents(name string) ModuleSource {
	data, err := os.ReadFile(name)

	if err != nil {
		panic(err)
	}

	return ModuleSource{
		Name:   name,
		Source: string(data),
	}
}

func TestVM(t *testing.T) {
	vm := NewSupervisor()
	iso := vm.NewIsolate()

	mainMod := NewCachedModule(vm, "main", srcFileContents(srcFileName))
	depMod := NewCachedModule(vm, "./vendors-node_modules_react_index_js.js", srcFileContents(srcFilename2))

	iso.moduleCache.Add(mainMod)
	iso.moduleCache.Add(depMod)

	ctx := NewContext(context.Background(), iso, nil)

	lm, err := ctx.Require(context.Background(), "main")

	require.NoError(t, err)

	if err := lm.register(); err != nil {
		panic(err)
	}
}
