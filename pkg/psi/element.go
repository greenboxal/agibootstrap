package psi

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"

	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/ast/astutil"
)

type Node interface {
	Node() ast.Node
	Parent() *Container

	IsContainer() bool
	IsLeaf() bool

	Comments() []*ast.Comment

	setParent(parent *Container)
}

type Container struct {
	node     ast.Node
	comments []*ast.Comment
	children []Node
	parent   *Container
}

func (c *Container) setParent(parent *Container) {
	if c.parent != nil {
		c.parent.removeChild(c)
	}

	c.parent = parent

	if c.parent != nil {
		c.parent.addChild(c)
	}
}

func (c *Container) IsContainer() bool        { return true }
func (c *Container) IsLeaf() bool             { return false }
func (c *Container) Comments() []*ast.Comment { return c.comments }
func (c *Container) Parent() *Container       { return c.parent }
func (c *Container) Node() ast.Node           { return c.node }
func (c *Container) Children() []Node         { return c.children }

func (c *Container) addChild(n Node) {
	idx := slices.Index(c.children, n)

	if idx != -1 {
		return
	}

	c.children = append(c.children, n)
}

func (c *Container) removeChild(n Node) {
	idx := slices.Index(c.children, n)

	if idx == -1 {
		return
	}

	c.children = slices.Delete(c.children, idx, idx+1)
}

func (sf *SourceFile) ToCode(node Node) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, sf.fset, node.Node())
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type Leaf struct {
	node   ast.Node
	parent *Container
}

func (f *Leaf) setParent(parent *Container) {
	if f.parent != nil {
		f.parent.removeChild(f)
	}

	f.parent = parent

	if f.parent != nil {
		f.parent.addChild(f)
	}
}

func (f *Leaf) IsContainer() bool        { return false }
func (f *Leaf) IsLeaf() bool             { return true }
func (f *Leaf) Parent() *Container       { return f.parent }
func (f *Leaf) Node() ast.Node           { return f.node }
func (f *Leaf) Comments() []*ast.Comment { return nil }

type SourceFile struct {
	name   string
	fset   *token.FileSet
	parsed *ast.File
	err    error

	root *Container
}

func (sf *SourceFile) FileSet() *token.FileSet { return sf.fset }
func (sf *SourceFile) Root() *Container        { return sf.root }
func (sf *SourceFile) Error() interface{}      { return sf.err }

type Package struct {
	Name    string
	Sources []*SourceFile
}

func Parse(filename string, sourceCode string) *SourceFile {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filename, sourceCode, parser.ParseComments)

	sf := &SourceFile{
		name:   filename,
		fset:   fset,
		parsed: parsed,
		err:    err,
	}

	if parsed != nil {
		sf.root = convertNode(parsed, sf).(*Container)
	}

	return sf
}

func convertNode(root ast.Node, sf *SourceFile) (result Node) {
	containerStack := stack.NewStack[*Container](16)

	astutil.Apply(root, func(cursor *astutil.Cursor) bool {
		node := cursor.Node()

		if node == nil {
			return false
		}

		_, parent := containerStack.Peek()

		wrapped := wrapNode(node)
		wrapped.setParent(parent)

		if c, ok := wrapped.(*Container); ok {
			for _, grp := range sf.parsed.Comments {
				if grp.Pos() >= node.Pos() && grp.End() <= node.End() {
					c.comments = append(c.comments, grp.List...)
				}
			}

			if err := containerStack.Push(c); err != nil {
				panic(err)
			}
		}

		return true
	}, func(cursor *astutil.Cursor) bool {
		node := cursor.Node()
		hasParent, parent := containerStack.Peek()

		if hasParent && parent.node == node {
			_, result = containerStack.Pop()
		}

		return true
	})

	return
}

func wrapNode(node ast.Node) Node {
	switch node.(type) {
	case *ast.File:
		return buildContainer(node)
	case *ast.FuncDecl:
		return buildContainer(node)
	case *ast.GenDecl:
		return buildContainer(node)
	case *ast.TypeSpec:
		return buildContainer(node)
	case *ast.ImportSpec:
		return buildContainer(node)
	case *ast.ValueSpec:
		return buildContainer(node)
	default:
		return buildLeaf(node)
	}
}

func buildContainer(node ast.Node) *Container {
	c := &Container{
		node: node,
	}

	return c
}

func buildLeaf(node ast.Node) *Leaf {
	return &Leaf{
		node: node,
	}
}

func cloneNode(node ast.Node) ast.Node {
	v := reflect.ValueOf(node)
	v = reflect.Indirect(v)
	clone := reflect.New(v.Type())

	clone.Elem().Set(v)

	return clone.Interface().(ast.Node)
}

func cloneTree(node ast.Node) ast.Node {
	cloned := cloneNode(node)
	v := reflect.ValueOf(cloned)
	t := v.Type()

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if f.Type.AssignableTo(reflect.TypeOf((*ast.Node)(nil)).Elem()) {
				v.Field(i).Set(reflect.ValueOf(cloneTree(v.Field(i).Interface().(ast.Node))))
			}
		}
	}

	return cloned
}
