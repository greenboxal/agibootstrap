package golang

import (
	"context"
	"go/parser"
	"go/token"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/codeanalysis"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, src *codeanalysis.SourceFile) (codeanalysis.CompilationUnit, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, src.Origin, src.Content, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	cu := &CompilationUnit{}
	cu.Root = GoAstToPsi(fset, parsed)

	if err := p.Analyze(ctx, cu); err != nil {
		return nil, err
	}

	return cu, nil
}

func (p *Parser) Analyze(ctx context.Context, cu *CompilationUnit) error {
	cu.Scope = codeanalysis.DefineNodeScope(cu.Root)
	cu.Scope.SetParent(cu.Root)

	if err := psi.Walk(cu.Root, cu.visitForAnalysis); err != nil {
		return err
	}

	return nil
}

type CompilationUnit struct {
	Root  psi.Node
	Scope *codeanalysis.Scope
}

func (u *CompilationUnit) GetRoot() psi.Node             { return u.Root }
func (u *CompilationUnit) GetScope() *codeanalysis.Scope { return u.Scope }

func (u *CompilationUnit) visitForAnalysis(cursor psi.Cursor, entering bool) error {
	n := cursor.Value()

	if entering {
		parentScope := codeanalysis.GetNodeHierarchyScope(n.Parent())

		switch n := n.(type) {
		case *FuncDecl:
			parentScope.DefineSymbol(n.GetName().Name).SetDeclaration(n)

			codeanalysis.DefineNodeScope(n)

		case *GenDecl:
			for _, spec := range n.GetSpecs() {
				switch spec := spec.(type) {
				case *ValueSpec:
					for _, name := range spec.GetNames() {
						parentScope.DefineSymbol(name.Name).SetDeclaration(n)
					}

				case *TypeSpec:
					parentScope.DefineSymbol(spec.GetName().Name).SetDeclaration(n)
				}
			}
		}
	} else {
		scope := codeanalysis.GetNodeHierarchyScope(n)

		switch n := n.(type) {
		case *Field:
			typ := n.GetType()

			if err := u.visitResolveReferences(scope, typ); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *CompilationUnit) visitResolveReferences(scope *codeanalysis.Scope, typ Expr) error {
	var names []string

	if typ == nil {
		return nil
	}

	for {
		switch n := typ.(type) {
		case *StarExpr:
			typ = n.GetX()
			continue

		case *SelectorExpr:
			names = slices.Insert(names, 0, n.GetSel().Name)
			typ = n.GetX()

		case *Ident:
			names = slices.Insert(names, 0, n.Name)

			sym := scope.ResolveSymbol(names...)

			if sym == nil {
				sym = scope.DefineSymbol(strings.Join(names, "."))
			}

			sym.AddReference(n)

			return nil

		default:
			panic("LOL")
		}
	}
}
