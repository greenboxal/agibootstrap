import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class OnNextTickRequest extends makeSchema("psidb.jukebox/OnNextTickRequest", {
    next_timecode: PrimitiveTypes.Float64,
}) {}
