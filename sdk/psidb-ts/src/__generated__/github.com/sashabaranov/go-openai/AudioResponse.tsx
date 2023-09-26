import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { struct } from "@psidb/psidb-sdk/types//struct";


export class AudioResponse extends makeSchema("github.com/sashabaranov/go-openai/AudioResponse", {
    duration: PrimitiveTypes.Float64,
    language: PrimitiveTypes.String,
    segments: ArrayOf(struct),
    task: PrimitiveTypes.String,
    text: PrimitiveTypes.String,
}) {}
