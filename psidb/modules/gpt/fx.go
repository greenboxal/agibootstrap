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

	inject.WithRegisteredService[indexing2.NodeEmbedder](inject.ServiceRegistrationScopeSingleton),
	inject.WithRegisteredService[*openai.Client](inject.ServiceRegistrationScopeSingleton),
	inject.WithRegisteredService[*EmbeddingCacheManager](inject.ServiceRegistrationScopeSingleton),
)
