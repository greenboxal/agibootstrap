package singularity

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type WorldState struct {
	psi.NodeBase

	KV map[string]any

	Epoch int64
	Cycle int64
	Step  int64

	SystemMessages []chat.Message
}

func NewWorldState() *WorldState {
	ws := &WorldState{
		KV: map[string]any{},
	}

	ws.Init(ws, "")

	return ws
}

func (w *WorldState) Get(key string) any {
	return w.KV[key]
}

func (w *WorldState) Set(key string, value any) {
	w.KV[key] = value
}
