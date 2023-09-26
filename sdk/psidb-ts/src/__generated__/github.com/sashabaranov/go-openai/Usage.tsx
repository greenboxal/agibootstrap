import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Usage extends makeSchema("github.com/sashabaranov/go-openai/Usage", {
    completion_tokens: PrimitiveTypes.Float64,
    prompt_tokens: PrimitiveTypes.Float64,
    total_tokens: PrimitiveTypes.Float64,
}) {}
