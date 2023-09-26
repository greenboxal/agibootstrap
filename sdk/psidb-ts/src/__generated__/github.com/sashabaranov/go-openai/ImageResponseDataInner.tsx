import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ImageResponseDataInner extends makeSchema("github.com/sashabaranov/go-openai/ImageResponseDataInner", {
    b64_json: PrimitiveTypes.String,
    url: PrimitiveTypes.String,
}) {}
