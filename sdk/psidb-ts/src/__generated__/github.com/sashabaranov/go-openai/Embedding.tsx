import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Embedding extends makeSchema("github.com/sashabaranov/go-openai/Embedding", {
    embedding: ArrayOf(PrimitiveTypes.Float32),
    index: PrimitiveTypes.Float64,
    object: PrimitiveTypes.String,
}) {}
