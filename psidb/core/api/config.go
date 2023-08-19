package coreapi

import (
	"os"
	"path"

	"github.com/alitto/pond"
)

type Config struct {
	RootUUID string

	ProjectDir        string
	DataDir           string
	EmbeddingCacheDir string

	ListenEndpoint string

	UseTLS      bool
	TLSCertFile string
	TLSKeyFile  string

	Workers ExecutorPoolConfig
}

func (cfg *Config) SetDefaults() {
	if cfg.ProjectDir == "" {
		cwd, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		cfg.ProjectDir = cwd
	}

	if cfg.DataDir == "" {
		cfg.DataDir = path.Join(cfg.ProjectDir, ".fti/psi")
	}

	if cfg.EmbeddingCacheDir == "" {
		cfg.EmbeddingCacheDir = os.Getenv("PSIDB_EMBEDDING_CACHE_DIR")
	}

	if cfg.EmbeddingCacheDir == "" {
		cfg.EmbeddingCacheDir = path.Join(cfg.DataDir, "embedding-cache")
	}
}

type ExecutorPoolConfig struct {
	MaxWorkers  int
	MaxCapacity int
}

func (wpc *ExecutorPoolConfig) Build() *pond.WorkerPool {
	return pond.New(wpc.MaxWorkers, wpc.MaxCapacity)
}
