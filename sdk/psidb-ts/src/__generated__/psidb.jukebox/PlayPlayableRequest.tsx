import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PlayPlayableRequest extends makeSchema("psidb.jukebox/PlayPlayableRequest", {
    channel: PrimitiveTypes.Float64,
}) {}
