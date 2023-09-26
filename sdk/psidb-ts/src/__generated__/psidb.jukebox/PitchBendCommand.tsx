import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PitchBendCommand extends makeSchema("psidb.jukebox/PitchBendCommand", {
    pitch_bend: PrimitiveTypes.Float64,
}) {}
