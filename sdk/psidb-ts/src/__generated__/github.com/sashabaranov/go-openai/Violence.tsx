import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Violence extends makeSchema("github.com/sashabaranov/go-openai/Violence", {
    filtered: PrimitiveTypes.Boolean,
    severity: PrimitiveTypes.String,
}) {}
