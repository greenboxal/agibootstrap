import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class Player extends makeSchema("psidb.jukebox/Player", {
    current_item: PrimitiveTypes.String,
    current_time_code: PrimitiveTypes.Float64,
    is_playing: PrimitiveTypes.Boolean,
    name: PrimitiveTypes.String,
    queue: ArrayOf(Path),
}) {}
