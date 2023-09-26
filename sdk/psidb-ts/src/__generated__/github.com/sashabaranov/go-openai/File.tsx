import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class File extends makeSchema("github.com/sashabaranov/go-openai/File", {
    bytes: PrimitiveTypes.Float64,
    created_at: PrimitiveTypes.Float64,
    filename: PrimitiveTypes.String,
    id: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
    owner: PrimitiveTypes.String,
    purpose: PrimitiveTypes.String,
}) {}
