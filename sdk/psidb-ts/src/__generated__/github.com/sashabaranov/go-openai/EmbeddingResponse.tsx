import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Embedding } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/Embedding";


export class EmbeddingResponse extends makeSchema("github.com/sashabaranov/go-openai/EmbeddingResponse", {
    data: ArrayOf(Embedding),
    model: PrimitiveTypes.Float64,
    object: PrimitiveTypes.String,
    usage: makeSchema("", {
        completion_tokens: PrimitiveTypes.Float64,
        prompt_tokens: PrimitiveTypes.Float64,
        total_tokens: PrimitiveTypes.Float64,
    }),
}) {}
