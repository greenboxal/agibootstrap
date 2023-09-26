import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class PlayCommandSheetRequest extends makeSchema("psidb.jukebox/PlayCommandSheetRequest", {
    command_script: PrimitiveTypes.String,
    command_scripts: ArrayOf(PrimitiveTypes.String),
}) {}
