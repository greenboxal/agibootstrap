package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

type Summary struct {
	Title   string `json:"title" jsonschema:"title=Title,description=Title of the message."`
	Summary string `json:"summary" jsonschema:"title=Summary,description=Summary of the message."`
}

func QuerySummary(ctx context.Context, history []*thoughtdb.Thought) (Summary, error) {
	res, _, err := Reflect[Summary](ctx, ReflectOptions{
		History: history,

		Query: "Summarize the chat history above.",

		ExampleInput: "The cat (Felis catus) is a domestic species of small carnivorous mammal.[1][2] It is the only domesticated species in the family Felidae and is commonly referred to as the domestic cat or house cat to distinguish it from the wild members of the family.[4] Cats are commonly kept as house pets but can also be farm cats or feral cats; the feral cat ranges freely and avoids human contact.[5] Domestic cats are valued by humans for companionship and their ability to kill vermin. About 60 cat breeds are recognized by various cat registries.[6]",

		ExampleOutput: Summary{
			Title:   "Cat",
			Summary: "The cat (Felis catus) is a small carnivorous mammal and the only domesticated species in the Felidae family. Often known as domestic or house cats, they are differentiated from their wild counterparts. Cats are typically kept as pets for companionship and their ability to kill pests, but they can also be found on farms or as feral cats that avoid humans. Approximately 60 cat breeds are recognized by various registries.",
		},
	})

	return res, err
}
