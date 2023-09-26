import { makeSchema, PrimitiveTypes, MapOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class CompletionRequest extends makeSchema("github.com/sashabaranov/go-openai/CompletionRequest", {
    best_of: PrimitiveTypes.Float64,
    echo: PrimitiveTypes.Boolean,
    frequency_penalty: PrimitiveTypes.Float64,
    logit_bias: MapOf(PrimitiveTypes.String, PrimitiveTypes.Integer),
    logprobs: PrimitiveTypes.Float64,
    max_tokens: PrimitiveTypes.Float64,
    model: PrimitiveTypes.String,
    n: PrimitiveTypes.Float64,
    presence_penalty: PrimitiveTypes.Float64,
    prompt: PrimitiveTypes.Any,
    stop: ArrayOf(PrimitiveTypes.String),
    stream: PrimitiveTypes.Boolean,
    suffix: PrimitiveTypes.String,
    temperature: PrimitiveTypes.Float64,
    top_p: PrimitiveTypes.Float64,
    user: PrimitiveTypes.String,
}) {}
