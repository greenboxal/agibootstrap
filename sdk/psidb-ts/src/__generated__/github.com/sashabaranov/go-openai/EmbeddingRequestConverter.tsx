import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { EmbeddingRequest } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/EmbeddingRequest";


export const EmbeddingRequestConverter = makeInterface({
    name: "github.com/sashabaranov/go-openai/EmbeddingRequestConverter",
    methods: {
        Convert: PrimitiveTypes.Func()(EmbeddingRequest),
    },
});
