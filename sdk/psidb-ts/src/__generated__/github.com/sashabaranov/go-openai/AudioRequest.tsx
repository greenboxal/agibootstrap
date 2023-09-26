import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Reader } from "@psidb/psidb-sdk/types/io/Reader";


export class AudioRequest extends makeSchema("github.com/sashabaranov/go-openai/AudioRequest", {
    FilePath: PrimitiveTypes.String,
    Format: PrimitiveTypes.String,
    Language: PrimitiveTypes.String,
    Model: PrimitiveTypes.String,
    Prompt: PrimitiveTypes.String,
    Reader: Reader,
    Temperature: PrimitiveTypes.Float64,
}) {}
