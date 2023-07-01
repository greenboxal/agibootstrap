package featureextractors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

type SpeakerPrediction struct {
	NextSpeaker string `json:"NextSpeaker" jsonschema:"description=The next speaker"`
}

func PredictNextSpeaker(ctx context.Context, plan Plan, history ...*thoughtdb.Thought) (SpeakerPrediction, error) {
	planJson, err := json.Marshal(plan)

	if err != nil {
		return SpeakerPrediction{}, err
	}

	res, _, err := Reflect[SpeakerPrediction](ctx, ReflectOptions{
		History: history,

		Query: fmt.Sprintf("**Current Plan:**\n```json\n%s\n```\nPredict who should be the next speaker in the chat history above so they get closer to reaching their goals. Choose PAIR as the next speaker if the current one seems stuck or unable to continue with the current plan.", planJson),
	})

	return res, err
}
