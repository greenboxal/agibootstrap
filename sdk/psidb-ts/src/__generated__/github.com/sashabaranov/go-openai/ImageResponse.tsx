import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { ImageResponseDataInner } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ImageResponseDataInner";


export class ImageResponse extends makeSchema("github.com/sashabaranov/go-openai/ImageResponse", {
    created: PrimitiveTypes.Float64,
    data: ArrayOf(ImageResponseDataInner),
}) {}
