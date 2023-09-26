import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Hate extends makeSchema("github.com/sashabaranov/go-openai/Hate", {
    filtered: PrimitiveTypes.Boolean,
    severity: PrimitiveTypes.String,
}) {}
