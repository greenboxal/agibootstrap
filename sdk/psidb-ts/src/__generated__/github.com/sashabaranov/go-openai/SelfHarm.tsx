import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SelfHarm extends makeSchema("github.com/sashabaranov/go-openai/SelfHarm", {
    filtered: PrimitiveTypes.Boolean,
    severity: PrimitiveTypes.String,
}) {}
