package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/agents"
)

type SentimentScore struct {
	Anger     float64 `json:"anger"`
	Fear      float64 `json:"fear"`
	Joy       float64 `json:"joy"`
	Sadness   float64 `json:"sadness"`
	Analytic  float64 `json:"analytic"`
	Confident float64 `json:"confident"`
	Tentative float64 `json:"tentative"`
}

func QuerySentimentScore(ctx context.Context, history []agents.Message) (SentimentScore, error) {
	res, _, err := Reflect[SentimentScore](ctx, ReflectOptions{
		History: history,

		Query: "Please score each emotion present on the chat history above from 0 to 1.",

		ExampleInput: "i feel very honoured to be included in a magazine which prioritises health and clean living so highly...",
	})

	return res, err
}
