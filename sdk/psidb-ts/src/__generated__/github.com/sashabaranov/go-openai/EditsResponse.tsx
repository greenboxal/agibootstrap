import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { EditsChoice } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/EditsChoice";


export class EditsResponse extends makeSchema("github.com/sashabaranov/go-openai/EditsResponse", {
    choices: ArrayOf(EditsChoice),
    created: PrimitiveTypes.Float64,
    object: PrimitiveTypes.String,
    usage: makeSchema("", {
        completion_tokens: PrimitiveTypes.Float64,
        prompt_tokens: PrimitiveTypes.Float64,
        total_tokens: PrimitiveTypes.Float64,
    }),
}) {}
