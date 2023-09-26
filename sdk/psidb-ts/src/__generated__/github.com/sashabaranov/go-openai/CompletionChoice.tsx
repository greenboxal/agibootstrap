import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { map } from "@psidb/psidb-sdk/types//map";


export class CompletionChoice extends makeSchema("github.com/sashabaranov/go-openai/CompletionChoice", {
    finish_reason: PrimitiveTypes.String,
    index: PrimitiveTypes.Float64,
    logprobs: makeSchema("", {
        text_offset: ArrayOf(PrimitiveTypes.Integer),
        token_logprobs: ArrayOf(PrimitiveTypes.Float32),
        tokens: ArrayOf(PrimitiveTypes.String),
        top_logprobs: ArrayOf(map(PrimitiveTypes.String, PrimitiveTypes.Float32)),
    }),
    text: PrimitiveTypes.String,
}) {}
