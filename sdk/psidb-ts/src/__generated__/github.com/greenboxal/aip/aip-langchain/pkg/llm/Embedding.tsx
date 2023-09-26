import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Embedding extends makeSchema("github.com/greenboxal/aip/aip-langchain/pkg/llm/Embedding", {
    Embeddings: ArrayOf(PrimitiveTypes.Float32),
}) {}
