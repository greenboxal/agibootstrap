import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class InstrumentConfig extends makeSchema("psidb.jukebox/InstrumentConfig", {
    channel: PrimitiveTypes.Float64,
}) {}
