import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class VolumeCommand extends makeSchema("psidb.jukebox/VolumeCommand", {
    volume: PrimitiveTypes.Float64,
}) {}
