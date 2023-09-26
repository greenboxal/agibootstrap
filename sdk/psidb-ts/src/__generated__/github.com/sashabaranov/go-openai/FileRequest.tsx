import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FileRequest extends makeSchema("github.com/sashabaranov/go-openai/FileRequest", {
    file: PrimitiveTypes.String,
    purpose: PrimitiveTypes.String,
}) {}
