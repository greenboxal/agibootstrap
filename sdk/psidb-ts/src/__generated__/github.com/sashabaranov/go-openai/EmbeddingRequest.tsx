import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class EmbeddingRequest extends makeSchema("github.com/sashabaranov/go-openai/EmbeddingRequest", {
    input: PrimitiveTypes.Any,
    model: PrimitiveTypes.Float64,
    user: PrimitiveTypes.String,
}) {}
