package golang

import (
	"bytes"
	"context"
	"go/parser"
	"go/token"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

const EdgeKindDeclarations psi.EdgeKind = "Decls"

type SourceFile struct {
	psi.NodeBase

	name   string
	handle repofs.FileHandle

	l    *Language
	fset *token.FileSet
	dec  *decorator.Decorator

	root   psi.Node
	parsed *dst.File
	err    error

	original string
}

func NewSourceFile(l *Language, name string, handle repofs.FileHandle) *SourceFile {
	sf := &SourceFile{
		l:    l,
		fset: l.project.FileSet(),

		name:   name,
		handle: handle,
	}

	sf.dec = decorator.NewDecorator(sf.fset)

	sf.Init(sf, sf.name)

	return sf
}

func (sf *SourceFile) Name() string                    { return sf.name }
func (sf *SourceFile) Language() psi.Language          { return sf.l }
func (sf *SourceFile) Decorator() *decorator.Decorator { return sf.dec }
func (sf *SourceFile) Path() string                    { return sf.name }
func (sf *SourceFile) FileSet() *token.FileSet         { return sf.fset }
func (sf *SourceFile) OriginalText() string            { return sf.original }
func (sf *SourceFile) Root() psi.Node                  { return sf.root }
func (sf *SourceFile) Error() error                    { return sf.err }

func (sf *SourceFile) Load() error {
	file, err := sf.handle.Get()

	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	sf.root = nil
	sf.parsed = nil
	sf.err = nil
	sf.original = string(data)

	_, err = sf.Parse(sf.name, string(data))

	sf.err = err

	return err
}

func (sf *SourceFile) Replace(code string) error {
	if code == sf.original {
		return nil
	}

	err := sf.handle.Put(bytes.NewBufferString(code))

	if err != nil {
		return err
	}

	return sf.Load()
}

func (sf *SourceFile) SetRoot(node *dst.File) error {
	sf.parsed = node
	sf.root = AstToPsi(sf.parsed)
	sf.root.SetParent(sf)

	return nil
}

func (sf *SourceFile) Parse(filename string, sourceCode string) (result psi.Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = r.(error)
			}

			err = errors.Wrap(err, "panic while parsing file: "+filename)
		}
	}()

	parsed, err := decorator.ParseFile(sf.fset, filename, sourceCode, parser.ParseComments)

	sf.parsed = parsed
	sf.err = err

	if parsed == nil {
		return nil, err
	}

	node := AstToPsi(parsed)

	if sf.root == nil {
		if err := sf.SetRoot(parsed); err != nil {
			return nil, err
		}
	}

	return node, err
}

func (sf *SourceFile) ToCode(node psi.Node) (mdutils.CodeBlock, error) {
	var buf bytes.Buffer

	n := node.(Node).Ast()
	f, ok := n.(*dst.File)

	if !ok {
		var decls []dst.Decl

		obj := &dst.Object{}

		switch n := n.(type) {
		case *dst.FuncDecl:
			obj.Kind = dst.Fun

			decls = append(decls, n)

		case *dst.GenDecl:
			switch n.Tok {
			case token.CONST:
				obj.Kind = dst.Con
			case token.TYPE:
				obj.Kind = dst.Typ
			case token.VAR:
				obj.Kind = dst.Var
			}

			decls = append(decls, n)

		case *dst.TypeSpec:
			obj.Kind = dst.Typ

			decls = append(decls, &dst.GenDecl{
				Tok:   token.TYPE,
				Specs: []dst.Spec{n},
			})

		default:
			return mdutils.CodeBlock{}, errors.New("node is not a file or decl")
		}

		f = &dst.File{
			Name:       sf.parsed.Name,
			Decls:      decls,
			Scope:      dst.NewScope(nil),
			Imports:    sf.parsed.Imports,
			Unresolved: sf.parsed.Unresolved,
			Decs:       sf.parsed.Decs,
		}

		f.Scope.Insert(obj)
	}

	err := decorator.Fprint(&buf, f)

	if err != nil {
		return mdutils.CodeBlock{}, err
	}

	return mdutils.CodeBlock{
		Language: string(LanguageID),
		Code:     buf.String(),
		Filename: sf.Name(),
	}, nil
}

// MergeCompletionResults merges the completion results of a code block into the NodeProcessor.
// It takes a context.Context, a *NodeScope representing the current scope, a *psi.Cursor representing the current cursor position, and a psi.Node representing the completion results.
// This function merges the completion results into the AST being processed in the NodeProcessor by performing the following steps:
// 1. Merge the completion results into the current AST by calling the MergeFiles function.
// 2. Iterate over each declaration in the completion results.
// 3. Check if the declaration is a function and if its name matches the name of the current scope's function.
// 4. If the declaration matches, replace the current declaration at the cursor position with the new declaration by calling the ReplaceDeclarationAt function.
// 5. If the declaration doesn't match, merge the new declaration with the existing declarations by calling the MergeDeclarations function.
// 6. Return nil, indicating that there were no errors during the merging process.
func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope psi.Scope, cursor psi.Cursor, newAst psi.Node) error {
	MergeFiles(sf.root.(Node).Ast().(*dst.File), newAst.(Node).Ast().(*dst.File))

	for _, decl := range newAst.Children() {
		if funcType, ok := decl.(Node).Ast().(*dst.FuncDecl); ok && funcType.Name.Name == scope.Root().(Node).Ast().(*dst.FuncDecl).Name.Name {
			sf.ReplaceDeclarationAt(cursor, decl, funcType.Name.Name)
		} else {
			sf.MergeDeclarations(cursor, decl)
		}
	}

	return nil
}

// MergeDeclarations merges the declarations of a node into the NodeProcessor.
// It takes a psi.Cursor and a psi.Node representing the current node, and returns a boolean value indicating whether the merging was successful or not.
//
// The process of merging declarations involves the following steps:
// 1. Retrieve all declaration names from the node using the getDeclarationNames function.
// 2. For each declaration name, check if it already exists in the NodeProcessor. If not, insert the declaration at the cursor position using the InsertDeclarationAt function.
// 3. If the declaration already exists, check if the current cursor node is the same as the previous node. If so, replace the previous node with the current node's AST representation using the cursor.Replace function.
// 4. Update the existing declaration with the current node's index, name, and AST representation using the setExistingDeclaration function.
//
// The purpose of MergeDeclarations is to ensure that all declarations within a node are properly merged into the NodeProcessor, allowing further processing and code generation to be performed accurately.
func (sf *SourceFile) MergeDeclarations(cursor psi.Cursor, node psi.Node) bool {
	names := getDeclarationNames(node)

	for _, name := range names {
		previous := sf.Root().GetEdge(psi.EdgeKey{Kind: EdgeKindDeclarations, Name: name})

		if previous == nil {
			sf.InsertDeclarationAt(cursor, name, node)
		} else {
			if cursor.Node().(Node).Ast() == previous.To().(Node).Ast() {
				cursor.Replace(node)
			}

			sf.setExistingDeclaration(previous.Key().Index, name, node)
		}
	}

	return true
}

// InsertDeclarationAt inserts a declaration after the given cursor.
// It takes a psi.Cursor, a name string, and a decl psi.Node.
// The process involves the following steps:
// 1. Calling the InsertAfter method of the cursor and passing decl.Ast() to insert the declaration after the cursor.
// 2. Getting the current index of decl in the root file's declarations using the slices.Index method.
// 3. Calling the setExistingDeclaration method of the NodeProcessor to update the existing declaration information.
//
// The purpose of InsertDeclarationAt is to insert a declaration at a specific position in the AST and update the declaration information in the NodeProcessor for further processing and code generation.
func (sf *SourceFile) InsertDeclarationAt(cursor psi.Cursor, name string, decl psi.Node) {
	cursor.InsertAfter(decl)
	index := slices.Index(sf.root.(Node).Ast().(*dst.File).Decls, decl.(Node).Ast().(dst.Decl))
	sf.setExistingDeclaration(index, name, decl)
}

// ReplaceDeclarationAt method replaces a declaration at a specific cursor position with a new declaration.
//
// The ReplaceDeclarationAt method takes three parameters: a psi.Cursor, the declaration node to replace, and the name of the declaration.
//
// The steps involved in the ReplaceDeclarationAt method are as follows:
// 1. The method replaces the declaration node at the cursor position with the new declaration node using the cursor.Replace method.
// 2. It gets the index of the new declaration in the root file's declarations using the slices.Index method.
// 3. It updates the existing declaration information in the NodeProcessor by calling the setExistingDeclaration method.
//
// The purpose of the ReplaceDeclarationAt method is to provide a mechanism for replacing a declaration at a specific position in the AST and updating the declaration information in the NodeProcessor for further processing and code generation.
func (sf *SourceFile) ReplaceDeclarationAt(cursor psi.Cursor, decl psi.Node, name string) {
	cursor.Replace(decl)
	index := slices.Index(sf.root.(Node).Ast().(*dst.File).Decls, decl.(Node).Ast().(dst.Decl))
	sf.setExistingDeclaration(index, name, decl)
}

// setExistingDeclaration updates the information about an existing declaration.
// It takes an index, a name string, and a psi.Node representing the declaration.
// The process involves the following steps:
// 1. Retrieve the existing declaration from the NodeProcessor using the name.
// 2. If the declaration does not exist, create a new declaration and add it to the NodeProcessor.
// 3. Update the declaration's element, node, and index with the provided values.
//
// This function is responsible for maintaining and updating the information about existing declarations,
// ensuring their accuracy and consistency throughout the code generation process.
func (sf *SourceFile) setExistingDeclaration(index int, name string, node psi.Node) {
	sf.Root().SetEdge(psi.EdgeKey{Kind: EdgeKindDeclarations, Name: name, Index: index}, node)
}

func MergeFiles(file1, file2 *dst.File) *dst.File {
	mergedFile := file1
	newDecls := make([]dst.Decl, 0)

	for _, decl2 := range file2.Decls {
		found := false
		switch decl2 := decl2.(type) {
		case *dst.FuncDecl:
			for i, decl1 := range mergedFile.Decls {
				if decl1, ok := decl1.(*dst.FuncDecl); ok && decl1.Name.Name == decl2.Name.Name {
					mergedFile.Decls[i] = decl2
					found = true
					break
				}
			}
		case *dst.GenDecl:
			for _, spec2 := range decl2.Specs {
				switch spec2 := spec2.(type) {
				case *dst.TypeSpec:
					for i, decl1 := range mergedFile.Decls {
						if decl1, ok := decl1.(*dst.GenDecl); ok {
							for j, spec1 := range decl1.Specs {
								if spec1, ok := spec1.(*dst.TypeSpec); ok && spec1.Name.Name == spec2.Name.Name {
									decl1.Specs[j] = spec2
									found = true
									break
								}
							}
						}
						mergedFile.Decls[i] = decl1
					}
				case *dst.ValueSpec:
					for i, decl1 := range mergedFile.Decls {
						if decl1, ok := decl1.(*dst.GenDecl); ok {
							for j, spec1 := range decl1.Specs {
								if spec1, ok := spec1.(*dst.ValueSpec); ok && spec1.Names[0].Name == spec2.Names[0].Name {
									decl1.Specs[j] = spec2
									found = true
									break
								}
							}
						}
						mergedFile.Decls[i] = decl1
					}
				}
			}
		}

		if !found {
			newDecls = append(newDecls, decl2)
		}
	}

	mergedFile.Decls = append(mergedFile.Decls, newDecls...)

	return mergedFile
}

// getDeclarationNames returns a slice of strings representing the declaration names in the given PSI node.
// The node parameter should be a psi.Node representing the AST node being processed.
// The function iterates through the AST node and extracts the names of declarations such as constants, variables, types, and functions.
// The extracted names are then appended to the names slice and returned.
func getDeclarationNames(node psi.Node) []string {
	var names []string

	switch d := node.(Node).Ast().(type) {
	case *dst.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *dst.ValueSpec: // for constants and variables
				for _, name := range s.Names {
					names = append(names, name.Name)
				}
			case *dst.TypeSpec: // for types
				names = append(names, s.Name.Name)
			}
		}
	case *dst.FuncDecl: // for functions
		names = append(names, d.Name.Name)
	}

	return names
}
