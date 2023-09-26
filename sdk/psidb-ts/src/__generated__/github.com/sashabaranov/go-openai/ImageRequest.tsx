import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ImageRequest extends makeSchema("github.com/sashabaranov/go-openai/ImageRequest", {
    n: PrimitiveTypes.Float64,
    prompt: PrimitiveTypes.String,
    response_format: PrimitiveTypes.String,
    size: PrimitiveTypes.String,
    user: PrimitiveTypes.String,
}) {}
