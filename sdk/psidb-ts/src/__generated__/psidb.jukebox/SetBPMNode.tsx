import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SetBPMNode extends makeSchema("psidb.jukebox/SetBPMNode", {
    BPM: PrimitiveTypes.Float64,
}) {}
