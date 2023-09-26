import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class VolumeNode extends makeSchema("psidb.jukebox/VolumeNode", {
    Volume: PrimitiveTypes.Float64,
}) {}
