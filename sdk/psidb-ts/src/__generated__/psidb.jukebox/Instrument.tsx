import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Instrument extends makeSchema("psidb.jukebox/Instrument", {
    config: makeSchema("", {
        channel: PrimitiveTypes.Float64,
    }),
    is_playing: PrimitiveTypes.Boolean,
    last_timecode: PrimitiveTypes.Float64,
    name: PrimitiveTypes.String,
}) {}
