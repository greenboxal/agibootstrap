import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PlayNoteNode extends makeSchema("psidb.jukebox/PlayNoteNode", {
    Accidentals: ArrayOf(PrimitiveTypes.String),
    Duration: PrimitiveTypes.Float64,
    Note: PrimitiveTypes.String,
    Octave: PrimitiveTypes.Float64,
    Velocity: PrimitiveTypes.Float64,
}) {}
