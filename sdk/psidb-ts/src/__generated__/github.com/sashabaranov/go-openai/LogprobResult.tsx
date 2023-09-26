import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { map } from "@psidb/psidb-sdk/types//map";


export class LogprobResult extends makeSchema("github.com/sashabaranov/go-openai/LogprobResult", {
    text_offset: ArrayOf(PrimitiveTypes.Integer),
    token_logprobs: ArrayOf(PrimitiveTypes.Float32),
    tokens: ArrayOf(PrimitiveTypes.String),
    top_logprobs: ArrayOf(map(PrimitiveTypes.String, PrimitiveTypes.Float32)),
}) {}
