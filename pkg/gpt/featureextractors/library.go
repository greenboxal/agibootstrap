package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Library struct {
	psi.NodeBase

	Books []Book `json:"books" jsonschema:"title=Books,description=Books in the library."`
}

type Book struct {
	psi.NodeBase

	Title    string    `json:"title" jsonschema:"title=Title,description=Title of the book."`
	Author   string    `json:"author" jsonschema:"title=Author,description=Author of the book."`
	Chapters []Chapter `json:"chapters" jsonschema:"title=Chapters,description=Chapters of the book."`
}

type Chapter struct {
	psi.NodeBase

	Title       string `json:"title" jsonschema:"title=Title,description=Title of the chapter."`
	Description string `json:"body" jsonschema:"title=Body,description=Body of the chapter."`
	Pages       []Page `json:"pages" jsonschema:"title=Pages,description=Pages of the chapter."`
}

type Page struct {
	psi.NodeBase

	Body       string              `json:"body" jsonschema:"title=Body,description=Body of the page."`
	CodeBlocks []mdutils.CodeBlock `json:"codeBlocks" jsonschema:"title=CodeBlocks,description=Code blocks of the page."`
}

func QueryLibrary(ctx context.Context, history []*thoughtdb.Thought) (Library, error) {
	res, _, err := Reflect[Library](ctx, ReflectOptions{
		History: history,

		Query: "Extract the knowledge base from the chat history above in the form of books, chapters, and pages.",
	})

	return res, err
}
