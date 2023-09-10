package typesystem

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	gopath "path"
	"path/filepath"
	"strings"
)

// ExtractGoComments will read all the go files contained in the provided path,
// including sub-directories, in order to generate a dictionary of comments
// associated with Types and Fields. The results will be added to the `commentsMap`
// provided in the parameters and expected to be used for Schema "description" fields.
//
// The `go/parser` library is used to extract all the comments and unfortunately doesn't
// have a built-in way to determine the fully qualified name of a package. The `base` paremeter,
// the URL used to import that package, is thus required to be able to match reflected types.
//
// When parsing type comments, we use the `go/doc`'s Synopsis method to extract the first phrase
// only. Field comments, which tend to be much shorter, will include everything.
func ExtractGoComments(base, pkgPath string, commentMap map[string]string) error {
	fset := token.NewFileSet()
	dict := make(map[string][]*ast.Package)
	err := filepath.Walk(pkgPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			d, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			for _, v := range d {
				// paths may have multiple packages, like for tests
				relPath, _ := filepath.Rel(pkgPath, path)
				k := gopath.Join(base, relPath)
				dict[k] = append(dict[k], v)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for pkg, p := range dict {
		for _, f := range p {
			gtxt := ""
			typ := ""
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					name := x.Name.String()

					if ast.IsExported(name) {
						if x.Recv != nil {
							receiverName := ""
							receiverTyp := x.Recv.List[0].Type

							for {
								if id, ok := receiverTyp.(*ast.Ident); ok {
									receiverName = id.Name
									break
								}

								switch x := receiverTyp.(type) {
								case *ast.StarExpr:
									receiverTyp = x.X
								case *ast.IndexExpr:
									receiverTyp = x.X
								case *ast.IndexListExpr:
									receiverTyp = x.X
								default:
									panic(fmt.Errorf("unexpected receiver type %T", receiverTyp))
								}
							}

							if receiverName != "" {
								name = receiverName + "." + name
							}
						}

						commentMap[fmt.Sprintf("%s.%s", pkg, name)] = strings.TrimSpace(x.Doc.Text())
					}
				case *ast.TypeSpec:
					typ = x.Name.String()
					if !ast.IsExported(typ) {
						typ = ""
					} else {
						txt := x.Doc.Text()
						if txt == "" && gtxt != "" {
							txt = gtxt
							gtxt = ""
						}
						commentMap[fmt.Sprintf("%s.%s", pkg, typ)] = strings.TrimSpace(txt)
					}
				case *ast.Field:
					txt := x.Doc.Text()
					if typ != "" && txt != "" {
						for _, n := range x.Names {
							if ast.IsExported(n.String()) {
								k := fmt.Sprintf("%s.%s.%s", pkg, typ, n)
								commentMap[k] = strings.TrimSpace(txt)
							}
						}
					}
				case *ast.GenDecl:
					// remember for the next type
					gtxt = x.Doc.Text()
				}
				return true
			})
		}
	}

	return nil
}
