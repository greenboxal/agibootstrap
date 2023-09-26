import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { EvaluableCommand } from "@psidb/psidb-sdk/types/psidb.jukebox/EvaluableCommand";


export class Song extends makeSchema("psidb.jukebox/Song", {
    commands: ArrayOf(EvaluableCommand),
    name: PrimitiveTypes.String,
    script: PrimitiveTypes.String,
}) {}
