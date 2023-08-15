package kb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

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

type DocumentDeduplication struct {
	Title string `json:"title" jsonschema:"title=Title,description=Title of the document."`
	Index int    `json:"index" jsonschema:"title=Index,description=Index of the document in the list."`

	IsDuplicated      bool `json:"isDuplicated" jsonschema:"title=Is Duplicated,description=Whether the documents contain any duplicates related to the query."`
	IsSameSubject     bool `json:"isSameSubject" jsonschema:"title=Is Same Subject,description=Whether the documents are related to the same subject."`
	IsSameContent     bool `json:"isSameContent" jsonschema:"title=Is Same Content,description=Whether the documents are related to the same content."`
	IsSameAuthor      bool `json:"isSameAuthor" jsonschema:"title=Is Same Author,description=Whether the documents are related to the same author."`
	IsSamePerspective bool `json:"isSamePerspective" jsonschema:"title=Is Same Perspective,description=Whether the documents are related to the same perspective."`
	IsSameTopic       bool `json:"isSameTopic" jsonschema:"title=Is Same Topic,description=Whether the documents are related to the same topic."`
	IsSameHeading     bool `json:"isSameHeading" jsonschema:"title=Is Same Heading,description=Whether the documents are related to the same heading."`
	IsSynonym         bool `json:"isSynonym" jsonschema:"title=Is Synonym,description=Whether the documents are related to the same synonym."`
}

type DocumentDeduplicationCandidate struct {
	Index       int    `json:"Index"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

func QueryDocumentDeduplication(ctx context.Context, req *KnowledgeRequest, candidates ...*Document) (DocumentDeduplication, error) {
	var msgs []*thoughtdb.Thought

	var allCandidates = make([]DocumentDeduplicationCandidate, len(candidates)+1)

	allCandidates[0] = DocumentDeduplicationCandidate{
		Index:       0,
		Title:       req.Title,
		Description: req.Description,
	}

	for i, doc := range candidates {
		allCandidates[i+1] = DocumentDeduplicationCandidate{
			Index:       i + 1,
			Title:       doc.Title,
			Description: doc.Body,
		}
	}

	for _, doc := range allCandidates {
		data, err := json.Marshal(doc)

		if err != nil {
			return DocumentDeduplication{}, err
		}

		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: string(data),
		})
	}

	res, _, err := featureextractors.Reflect[DocumentDeduplication](ctx, featureextractors.ReflectOptions{
		History: msgs,
		Query: fmt.Sprintf(
			"You are specialized in de-duplicating documents based on their description. Are there are any documents in the list above that could be related to %s?",
			req.Title,
		),
		ExampleInput: "",
	})

	return res, err
}
