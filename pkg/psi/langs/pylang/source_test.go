package pylang

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/build/codegen"
	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type testEnv struct {
	Project  project.Project
	Language *Language
}

var testCodeSimple = `

def doHello():
	println("Hello")

def main():
	doHello()
	doWorld()

def doWorld():
	println(" world\n")
`

var testCodeMerge = `
# Case: Keep

def doHello(self):
	println("Hello")

# Case: Replace existing declaration
def main(self):
	doHello(self)
	doWorld(self)

	newHelloFn(self)

# Case: Keep
def doWorld(self):
	println(" world\n")

# Case: New named declaration is inserted at the end
def newHelloFn(self):
	println("Again!\n")
`

func setupTestProject(t *testing.T) testEnv {
	p, err := codex.NewProject(".")
	require.NoError(t, err)
	lang := NewLanguage(p)

	return testEnv{
		Project:  p,
		Language: lang,
	}
}

func TestSourceParse(t *testing.T) {
	env := setupTestProject(t)

	src := NewSourceFile(env.Language, "test.py", repofs.String(testCodeSimple))

	require.NoError(t, src.Load())
	require.NotNil(t, src)
}

func TestSourcePrint(t *testing.T) {
	env := setupTestProject(t)

	src := NewSourceFile(env.Language, "test.py", repofs.String(testCodeSimple))

	require.NoError(t, src.Load())
	require.NotNil(t, src)

	code, err := src.ToCode(src.Root())

	require.NoError(t, err)
	require.Equal(t, "python", code.Language)
	require.Equal(t, "test.py", code.Filename)
	require.Equal(t, testCodeSimple, code.Code)
}

func TestSourceMerge(t *testing.T) {
	env := setupTestProject(t)

	src1 := NewSourceFile(env.Language, "test.py", repofs.String(testCodeSimple))
	src2 := NewSourceFile(env.Language, "merge.py", repofs.String(testCodeMerge))

	require.NotNil(t, src1)
	require.NotNil(t, src2)

	require.NoError(t, src1.Load())
	require.NoError(t, src2.Load())

	require.NotNil(t, src1.Root())
	require.NotNil(t, src2.Root())

	c := psi.NewCursor()
	c.SetCurrent(src1.Root())

	scope := &codegen.NodeScope{
		Node: src1.Root(),
	}

	err := src1.MergeCompletionResults(context.Background(), scope, c, src2, src2.Root())

	require.NoError(t, err)

	code, err := src1.ToCode(src1.Root())

	require.NoError(t, err)
	require.Equal(t, "python", code.Language)
	require.Equal(t, "test.py", code.Filename)
	require.Equal(t, testCodeMerge, code.Code)
}
