package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"text/template"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type AstNodeEdge struct {
	Name     string
	Unique   bool
	Optional bool
	Type     string
	TypeName string

	ResolvedType *AstNodeType

	IsNode      bool
	IsInterface bool
}

func (e AstNodeEdge) String() string {
	return fmt.Sprintf("var EdgeKind%s = psi.DefineEdgeType[*%s](\"Go%s\")", e.Name, e.Type, e.Name)
}

type AstNodeType struct {
	Order int

	Name        string
	Edges       []AstNodeEdge
	Attributes  []AstNodeEdge
	IsInterface bool
	SkipEmit    bool
}

func (n AstNodeType) String() string {
	t, err := template.New("").Parse(`
{{- $ctx := . }}
{{- if .Type.IsInterface }}
type {{ .Type.Name }} interface {
	Node
}
{{- else }}
type {{ .Type.Name }} struct {
	NodeBase

{{- range .Type.Attributes }}
	{{ .Name }} {{ .Type }} ` + "`" + `json:"{{ .Name }}"` + "`" + `
{{- end }}
}

var {{ .Type.Name }}Type = psi.DefineNodeType[*{{ .Type.Name }}]()
{{ end }}

{{ range .Type.Edges }}
var EdgeKind{{ $ctx.Type.Name }}{{ .Name }} = psi.DefineEdgeType[{{ .Type }}]("Go{{ $ctx.Type.Name }}{{ .Name }}")
{{- end }}

{{- range .Type.Edges }}
{{- if .Unique }}
func (n *{{ $ctx.Type.Name }}) Get{{ .Name }}() {{ .Type }} { return psi.GetEdgeOrNil[{{ .Type }}](n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}.Singleton()) }
func (n *{{ $ctx.Type.Name }}) Set{{ .Name }}(node {{ .Type }}) { psi.UpdateEdge(n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}.Singleton(), node) }
{{- else }}
func (n *{{ $ctx.Type.Name }}) Get{{ .Name }}() []{{ .Type }} { return psi.GetEdges(n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}) }
func (n *{{ $ctx.Type.Name }}) Set{{ .Name }}(nodes []{{ .Type }}) { psi.UpdateEdges(n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}, nodes) }
{{- end }}
{{- end }}

{{- if .Type.IsInterface }}
func NewFrom{{ .Type.Name }}(fset *token.FileSet, node ast.{{ .Type.Name }}) {{ .Type.Name }} {
	return GoAstToPsi(fset, node).({{ .Type.Name }})
}
{{- else }}
func NewFrom{{ .Type.Name }}(fset *token.FileSet, node *ast.{{ .Type.Name }}) *{{ .Type.Name }} {
	n := &{{ .Type.Name }}{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType({{ .Type.Name }}Type))

	return n
}

func (n *{{ .Type.Name }}) CopyFromGoAst(fset *token.FileSet, src *ast.{{ .Type.Name }}) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()

{{- range .Type.Attributes }}
	n.{{ .Name }} = src.{{ .Name }}
{{- end }}

{{- range .Type.Edges }}
{{- if .Unique }}
	if src.{{ .Name }} != nil {
		tmp{{ .Name }} := NewFrom{{ .TypeName }}(fset, src.{{ .Name }})
		tmp{{ .Name }}.SetParent(n)
		psi.UpdateEdge(n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}.Singleton(), tmp{{ .Name }})
	}
{{ else }}
	for i, v := range src.{{ .Name }} {
		tmp{{ .Name }} := NewFrom{{ .TypeName }}(fset, v)
		tmp{{ .Name }}.SetParent(n)
		psi.UpdateEdge(n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}.Indexed(int64(i)), tmp{{ .Name }})
	}
{{ end }}
{{- end }}
}

func (n *{{ .Type.Name }}) ToGoAst() ast.Node { return n.ToGo{{ .Type.Name }}(nil) }

func (n *{{ .Type.Name }}) ToGo{{ .Type.Name }}(dst *ast.{{ .Type.Name }}) *ast.{{ .Type.Name }} {
	if dst == nil {
		dst = &ast.{{ .Type.Name }}{}
	}

{{- range .Type.Attributes }}
	dst.{{ .Name }} = n.{{ .Name }}
{{- end }}

{{- range .Type.Edges }}
{{- if .Unique }}
	tmp{{ .Name }} := psi.GetEdgeOrNil[{{ .Type }}](n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }}.Singleton())
	if tmp{{ .Name }} != nil {
		dst.{{ .Name }} = tmp{{ .Name }}.ToGoAst()
{{- if .IsInterface -}}
.(ast.{{ .TypeName }})
{{- else -}}
.(*ast.{{ .TypeName }})
{{- end }}
	}
{{ else }}
	tmp{{ .Name }} := psi.GetEdges(n, EdgeKind{{ $ctx.Type.Name }}{{ .Name }})
	dst.{{ .Name }} = make([]*ast.{{ .TypeName }}, len(tmp{{ .Name }}))
	for i, v := range tmp{{ .Name }} {
		dst.{{ .Name }}[i] = v.ToGoAst().(*ast.{{ .TypeName }})
	}
{{ end }}
{{- end }}

	return dst
}
{{- end }}
`)

	if err != nil {
		panic(err)
	}

	buffer := bytes.NewBuffer(nil)

	if err := t.Execute(buffer, struct {
		Type AstNodeType
	}{
		Type: n,
	}); err != nil {
		panic(err)
	}

	return buffer.String()
}

func main() {
	filename := os.Args[1]

	nodes, err := parseAST(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(`package golang

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)
`)

	for _, node := range nodes {
		if node.SkipEmit {
			continue
		}
		/*fmt.Println("AST Node:", node.Name)
		fmt.Println("Type:", node.Type)
		fmt.Println("Edges:")
		for _, edge := range node.Edges {
			fmt.Println("  Name:", edge.Name)
			fmt.Println("  Unique:", edge.Unique)
			fmt.Println("  Optional:", edge.Optional)
			fmt.Println("  Type:", edge.Type)
		}*/
		fmt.Println(node.String())
		fmt.Println()
	}

	t, err := template.New("").Parse(`
{{- $ctx := . }}

func GoAstToPsi(fset *token.FileSet, n ast.Node) Node {
	switch n := n.(type) {
{{- range .Types }}
{{- if not .SkipEmit }}
{{- if not .IsInterface }}
	case *ast.{{ .Name }}:
		return NewFrom{{ .Name }}(fset, n)
{{- end }}
{{- end }}
{{- end }}
	default:
		panic(fmt.Errorf("unknown node type %T", n))
	}
}
`)

	if err != nil {
		panic(err)
	}

	if err := t.Execute(os.Stdout, struct {
		Types []*AstNodeType
	}{
		Types: nodes,
	}); err != nil {
		panic(err)
	}
}

func parseAST(filename string) ([]*AstNodeType, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	nodes := map[string]*AstNodeType{}

	ast.Inspect(file, func(n ast.Node) bool {
		if typ, ok := n.(*ast.TypeSpec); ok {
			nodeType := AstNodeType{
				Name:  typ.Name.Name,
				Order: len(nodes),
			}

			if _, ok := typ.Type.(*ast.InterfaceType); ok {
				nodeType.IsInterface = true
			}

			if nodeType.Name == "Node" || nodeType.Name == "Package" {
				nodeType.SkipEmit = true
			}

			switch t := typ.Type.(type) {
			case *ast.StructType:
				for _, field := range t.Fields.List {
					edge := AstNodeEdge{
						Name:   field.Names[0].Name,
						Unique: true,
						IsNode: true,
					}

					typ := field.Type

					if t, ok := typ.(*ast.ArrayType); ok {
						typ = t.Elt
						edge.Unique = false
					}

					buf := bytes.NewBuffer(nil)

					if err := printer.Fprint(buf, fset, typ); err != nil {
						panic(err)
					}

					edge.Type = buf.String()

					if t, ok := typ.(*ast.StarExpr); ok {
						typ = t.X
						edge.Optional = true
					}

					if t, ok := typ.(*ast.Ident); ok {
						if t.Name[0] >= 'a' && t.Name[0] <= 'z' {
							edge.IsNode = false
						} else {
							edge.IsNode = true
						}

						edge.TypeName = t.Name
					}

					if edge.Type == "ChanDir" {
						edge.IsNode = false
					}

					if edge.Type == "token.Token" {
						edge.IsNode = false
					}

					if edge.Type == "token.Pos" {
						edge.IsNode = false
					}

					if edge.Type == "*Object" {
						continue
					}

					if edge.Type == "*Scope" {
						continue
					}

					if edge.Type == "token.Pos" {
						continue
					}

					if edge.IsNode {
						nodeType.Edges = append(nodeType.Edges, edge)
					} else {
						nodeType.Attributes = append(nodeType.Attributes, edge)
					}
				}

			case *ast.InterfaceType:

			default:
				nodeType.SkipEmit = true
			}

			nodes[nodeType.Name] = &nodeType
		}

		return true
	})

	fmt.Fprintf(os.Stderr, "Found %d nodes\n", len(nodes))

	for _, typ := range nodes {
		fmt.Fprintf(os.Stderr, "* %s\n", typ.Name)

		for _, edge := range typ.Edges {
			if other, ok := nodes[edge.TypeName]; ok {
				edge.ResolvedType = other
				edge.IsInterface = other.IsInterface
			}

			fmt.Fprintf(os.Stderr, "  - %s : %s (%v)\n", edge.Name, edge.TypeName, edge.IsInterface)
		}
	}

	result := maps.Values(nodes)

	slices.SortFunc(result, func(a, b *AstNodeType) bool {
		return a.Order < b.Order
	})

	return result, nil
}
