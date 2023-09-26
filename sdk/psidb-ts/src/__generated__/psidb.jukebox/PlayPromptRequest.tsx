import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PlayPromptRequest extends makeSchema("psidb.jukebox/PlayPromptRequest", {
    prompt: PrimitiveTypes.String,
    start_timecode: PrimitiveTypes.Float64,
}) {}
