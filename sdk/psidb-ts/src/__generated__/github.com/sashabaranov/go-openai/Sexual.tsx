import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Sexual extends makeSchema("github.com/sashabaranov/go-openai/Sexual", {
    filtered: PrimitiveTypes.Boolean,
    severity: PrimitiveTypes.String,
}) {}
