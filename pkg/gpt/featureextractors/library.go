package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/mdutils"
)

type Library struct {
	Books []Book `json:"books" jsonschema:"title=Books,description=Books in the library."`
}

type Book struct {
	Title    string    `json:"title" jsonschema:"title=Title,description=Title of the book."`
	Author   string    `json:"author" jsonschema:"title=Author,description=Author of the book."`
	Chapters []Chapter `json:"chapters" jsonschema:"title=Chapters,description=Chapters of the book."`
}

type Chapter struct {
	Title       string `json:"title" jsonschema:"title=Title,description=Title of the chapter."`
	Description string `json:"body" jsonschema:"title=Body,description=Body of the chapter."`
	Pages       []Page `json:"pages" jsonschema:"title=Pages,description=Pages of the chapter."`
}

type Page struct {
	Body       string              `json:"body" jsonschema:"title=Body,description=Body of the page."`
	CodeBlocks []mdutils.CodeBlock `json:"codeBlocks" jsonschema:"title=CodeBlocks,description=Code blocks of the page."`
}

func QueryLibrary(ctx context.Context, history []agents.Message) (Library, error) {
	res, _, err := Reflect[Library](ctx, ReflectOptions{
		History: history,

		Query: "Extract the knowledge base from the chat history above in the form of books, chapters, and pages.",
	})

	return res, err
}
