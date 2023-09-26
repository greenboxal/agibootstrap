import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FineTuneDeleteResponse extends makeSchema("github.com/sashabaranov/go-openai/FineTuneDeleteResponse", {
    deleted: PrimitiveTypes.Boolean,
    id: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
}) {}
