import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class NoteCommand extends makeSchema("psidb.jukebox/NoteCommand", {
    accidentals: ArrayOf(PrimitiveTypes.String),
    duration: PrimitiveTypes.Float64,
    note: PrimitiveTypes.String,
    octave: PrimitiveTypes.Float64,
    velocity: PrimitiveTypes.Float64,
}) {}
