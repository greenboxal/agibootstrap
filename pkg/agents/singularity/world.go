package singularity

func NewWorldState() *WorldState {
	return &WorldState{
		KV: map[string]any{},
	}
}

type WorldState struct {
	KV map[string]any

	Epoch int64
	Cycle int64
	Step  int64
}

func (w *WorldState) Get(key string) any {
	return w.KV[key]
}

func (w *WorldState) Set(key string, value any) {
	w.KV[key] = value
}
