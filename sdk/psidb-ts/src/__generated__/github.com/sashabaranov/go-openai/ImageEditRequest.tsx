import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { File } from "@psidb/psidb-sdk/types/os/File";


export class ImageEditRequest extends makeSchema("github.com/sashabaranov/go-openai/ImageEditRequest", {
    image: File,
    mask: File,
    n: PrimitiveTypes.Float64,
    prompt: PrimitiveTypes.String,
    response_format: PrimitiveTypes.String,
    size: PrimitiveTypes.String,
}) {}
