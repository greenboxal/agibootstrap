package golang

import (
	"bytes"
	"fmt"
	"go/scanner"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

var hasPackageRegex = regexp.MustCompile(`(?m)^.*package\s+([a-zA-Z0-9_]+)\n`)
var hasHtmlEscapeRegex = regexp.MustCompile(`&lt;|&gt;|&amp;|&quot;|&#[0-9]{2};`)

const LanguageID psi.LanguageID = "go"

func init() {
	project.RegisterLanguage(LanguageID, func(p project.Project) psi.Language {
		return NewLanguage(p)
	})
}

type Language struct {
	project project.Project
}

func NewLanguage(p project.Project) *Language {
	return &Language{
		project: p,
	}
}

func (l *Language) Name() psi.LanguageID {
	return LanguageID
}

func (l *Language) Extensions() []string {
	return []string{".go"}
}

func (l *Language) CreateSourceFile(fileName string, fileHandle repofs.FileHandle) psi.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(fileName string, code string) (psi.SourceFile, error) {
	f := l.CreateSourceFile(fileName, &BufferFileHandle{data: code})

	if err := f.Load(); err != nil {
		return nil, err
	}

	return f, nil
}

// ParseCodeBlock parses the code block and returns the resulting PSI node.
// This function unescapes HTML escape sequences, modifies the package declaration,
// and merges the resulting code with the existing AST.
// It also handles orphan snippets by wrapping them in a pseudo function.
func (l *Language) ParseCodeBlock(blockName string, block mdutils.CodeBlock) (psi.SourceFile, error) {
	// Unescape HTML escape sequences in the code block
	if hasHtmlEscapeRegex.MatchString(block.Code) {
		block.Code = html.UnescapeString(block.Code)
	}

	patchedCode := block.Code
	pkgIndex := hasPackageRegex.FindStringIndex(patchedCode)

	if len(pkgIndex) > 0 {
		patchedCode = fmt.Sprintf("%s\n%s%s", patchedCode[:pkgIndex[1]], "\n", patchedCode[pkgIndex[1]:])
	} else {
		patchedCode = fmt.Sprintf("package gptimport\n%s", patchedCode)
	}

	patchedCode = hasPackageRegex.ReplaceAllString(patchedCode, "package gptimport\n")

	newRoot, e := l.Parse(blockName, patchedCode)

	if e != nil {
		if errList, ok := e.(scanner.ErrorList); ok {
			if len(errList) == 1 && strings.HasPrefix(errList[0].Msg, "expected declaration, ") {
				// Handle orphan snippets by wrapping them in a pseudo function
				patchedCode = fmt.Sprintf("package gptimport_orphan\nfunc orphanSnippet() {\n%s\n}\n", block.Code)
				newRoot2, e2 := l.Parse(blockName, patchedCode)

				if e2 != nil {
					return nil, e
				}

				newRoot = newRoot2
			}
		} else if e != nil {
			return nil, e
		}
	}

	return newRoot, nil
}

type BufferFileHandle struct {
	data string
}

type closerReader struct {
	io.Reader
}

func (c closerReader) Close() error {
	return nil
}

func (b BufferFileHandle) Get() (io.ReadCloser, error) {
	return closerReader{bytes.NewBufferString(b.data)}, nil
}

func (b BufferFileHandle) Put(src io.Reader) error {
	return errors.New("cannot put to buffer file handle")
}

func (b BufferFileHandle) Close() error {
	return nil
}
