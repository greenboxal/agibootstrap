import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Engine extends makeSchema("github.com/sashabaranov/go-openai/Engine", {
    id: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
    owner: PrimitiveTypes.String,
    ready: PrimitiveTypes.Boolean,
}) {}
