import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { File } from "@psidb/psidb-sdk/types/os/File";


export class ImageVariRequest extends makeSchema("github.com/sashabaranov/go-openai/ImageVariRequest", {
    image: File,
    n: PrimitiveTypes.Float64,
    response_format: PrimitiveTypes.String,
    size: PrimitiveTypes.String,
}) {}
