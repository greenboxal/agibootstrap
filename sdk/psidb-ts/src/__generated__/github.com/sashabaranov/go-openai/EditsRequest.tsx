import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class EditsRequest extends makeSchema("github.com/sashabaranov/go-openai/EditsRequest", {
    input: PrimitiveTypes.String,
    instruction: PrimitiveTypes.String,
    model: PrimitiveTypes.String,
    n: PrimitiveTypes.Float64,
    temperature: PrimitiveTypes.Float64,
    top_p: PrimitiveTypes.Float64,
}) {}
