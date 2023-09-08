package golang

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/legacy/build/codegen"
	"github.com/greenboxal/agibootstrap/pkg/legacy/codex"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type testEnv struct {
	Project  project.Project
	Language *Language
}

var testCodeSimple = `package main

func doHello() {
	println("Hello")
}

func main() {
	doHello()
	doWorld()
}

func doWorld() {
	println(" world\n")
}
`

var testCodeMerge = `package main

// Case: Keep
func doHello() {
	println("Hello")
}

// Case: Replace existing declaration
func main() {
	doHello()
	doWorld()

	newHelloFn()
}

// Case: Keep
func doWorld() {
	println(" world\n")
}

// Case: New named declaration is inserted at the end
func newHelloFn() {
	println("Again!\n")
}
`

func setupTestProject(t *testing.T) testEnv {
	p, err := codex.LoadProject(context.Background(), ".")
	require.NoError(t, err)
	lang := NewLanguage(p)

	return testEnv{
		Project:  p,
		Language: lang,
	}
}

func TestSourceParse(t *testing.T) {
	env := setupTestProject(t)

	src := NewSourceFile(env.Language, "test.go", repofs.String(testCodeSimple))

	require.NoError(t, src.Load())
	require.NotNil(t, src)
}

func TestSourcePrint(t *testing.T) {
	env := setupTestProject(t)

	src := NewSourceFile(env.Language, "test.go", repofs.String(testCodeSimple))

	require.NoError(t, src.Load())
	require.NotNil(t, src)

	code, err := src.ToCode(src.Root())

	require.NoError(t, err)
	require.Equal(t, "go", code.Language)
	require.Equal(t, "test.go", code.Filename)
	require.Equal(t, testCodeSimple, code.Code)
}

func TestSourceMerge(t *testing.T) {
	env := setupTestProject(t)

	src1 := NewSourceFile(env.Language, "test.go", repofs.String(testCodeSimple))
	src2 := NewSourceFile(env.Language, "merge.go", repofs.String(testCodeMerge))

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
	require.Equal(t, "go", code.Language)
	require.Equal(t, "test.go", code.Filename)
	require.Equal(t, testCodeMerge, code.Code)
}
