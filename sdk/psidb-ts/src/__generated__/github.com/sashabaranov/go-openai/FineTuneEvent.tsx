import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FineTuneEvent extends makeSchema("github.com/sashabaranov/go-openai/FineTuneEvent", {
    created_at: PrimitiveTypes.Float64,
    level: PrimitiveTypes.String,
    message: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
}) {}
