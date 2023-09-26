import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class EditsChoice extends makeSchema("github.com/sashabaranov/go-openai/EditsChoice", {
    index: PrimitiveTypes.Float64,
    text: PrimitiveTypes.String,
}) {}
