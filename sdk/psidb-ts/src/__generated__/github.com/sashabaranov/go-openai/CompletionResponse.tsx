import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { CompletionChoice } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/CompletionChoice";


export class CompletionResponse extends makeSchema("github.com/sashabaranov/go-openai/CompletionResponse", {
    choices: ArrayOf(CompletionChoice),
    created: PrimitiveTypes.Float64,
    id: PrimitiveTypes.String,
    model: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
    usage: makeSchema("", {
        completion_tokens: PrimitiveTypes.Float64,
        prompt_tokens: PrimitiveTypes.Float64,
        total_tokens: PrimitiveTypes.Float64,
    }),
}) {}
