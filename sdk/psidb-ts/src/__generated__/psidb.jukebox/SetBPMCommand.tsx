import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SetBPMCommand extends makeSchema("psidb.jukebox/SetBPMCommand", {
    bpm: PrimitiveTypes.Float64,
}) {}
