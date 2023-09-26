import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SongRegenerateRequest extends makeSchema("psidb.jukebox/SongRegenerateRequest", {
    prompt: PrimitiveTypes.String,
}) {}
