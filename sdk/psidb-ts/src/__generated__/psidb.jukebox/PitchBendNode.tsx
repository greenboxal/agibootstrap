import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PitchBendNode extends makeSchema("psidb.jukebox/PitchBendNode", {
    PitchBend: PrimitiveTypes.Float64,
}) {}
