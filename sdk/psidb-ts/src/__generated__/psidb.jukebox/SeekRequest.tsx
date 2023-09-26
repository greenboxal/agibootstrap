import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SeekRequest extends makeSchema("psidb.jukebox/SeekRequest", {
    time: PrimitiveTypes.Float64,
}) {}
