package graphfs

type Frame struct {
	Log []JournalEntry `json:"log"`
}
