import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ModerationRequest extends makeSchema("github.com/sashabaranov/go-openai/ModerationRequest", {
    input: PrimitiveTypes.String,
    model: PrimitiveTypes.String,
}) {}
