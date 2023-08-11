package kb

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

type DocumentRelatedContent struct {
	Related []*DocumentRelatedContentEntry `json:"related" jsonschema:"title=Related Items,description=List of related topics and its descriptions."`
}

type DocumentRelatedContentEntry struct {
	Title       string `json:"title" jsonschema:"title=Title,description=Title of the related content."`
	Description string `json:"description" jsonschema:"title=Description,description=Description of the related content."`
}

func QueryDocumentRelatedContent(ctx context.Context, history []*thoughtdb.Thought) (DocumentRelatedContent, error) {
	res, _, err := featureextractors.Reflect[DocumentRelatedContent](ctx, featureextractors.ReflectOptions{
		History: history,
		Query:   "Write a list of related topics and its descriptions.",

		ExampleInput: "",
	})

	return res, err
}

type DocumentCategories struct {
	Categories []string `json:"categories" jsonschema:"title=Categories,description=List of categories for this article."`
}

func QueryDocumentCategories(ctx context.Context, history []*thoughtdb.Thought) (DocumentCategories, error) {
	res, _, err := featureextractors.Reflect[DocumentCategories](ctx, featureextractors.ReflectOptions{
		History: history,
		Query:   "Categorize the document above.",

		ExampleInput: "",
	})

	return res, err
}
