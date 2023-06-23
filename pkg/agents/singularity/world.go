package singularity

func newWorldState() *worldState {
	return &worldState{
		KV: map[string]any{},
	}
}

type worldState struct {
	KV map[string]any

	Epoch int64
	Cycle int64
	Step  int64
}

func (w *worldState) Get(key string) any {
	return w.KV[key]
}

func (w *worldState) Set(key string, value any) {
	w.KV[key] = value
}
