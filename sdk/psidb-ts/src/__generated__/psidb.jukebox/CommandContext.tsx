import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class CommandContext extends makeSchema("psidb.jukebox/CommandContext", {
    BPM: PrimitiveTypes.Float64,
    channel: PrimitiveTypes.Float64,
}) {}
