package fti

var defaultConfig = Config{
	Embedding: struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}{
		Provider: "OpenAI",
		Model:    "AdaEmbeddingV2",
	},
	ChunkSpecs: []ChunkSpec{
		{MaxTokens: 512, Overlap: 128},
		{MaxTokens: 1024, Overlap: 256},
	},
}

type Config struct {
	Embedding struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	} `json:"embedding"`

	ChunkSpecs []ChunkSpec `json:"chunk_specs"`
}

type ChunkSpec struct {
	MaxTokens int `json:"max_tokens"`
	Overlap   int `json:"overlap"`
}
