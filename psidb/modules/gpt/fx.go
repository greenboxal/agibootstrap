package gpt

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	indexing2 "github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

var Module = fx.Module(
	"modules/gpt",

	fx.Provide(NewEmbeddingCacheManager),

	fx.Provide(func() *openai.Client {
		return GlobalClient
	}),

	fx.Invoke(func(sp inject.ServiceProvider, e indexing2.NodeEmbedder, client *openai.Client) {
		inject.RegisterInstance(sp, e)
		inject.RegisterInstance(sp, client)
	}),
)
