import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Result } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/Result";


export class ModerationResponse extends makeSchema("github.com/sashabaranov/go-openai/ModerationResponse", {
    id: PrimitiveTypes.String,
    model: PrimitiveTypes.String,
    results: ArrayOf(Result),
}) {}
