import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class IndexEntryChunk extends makeSchema("psidb.docs/IndexEntryChunk", {
    C: PrimitiveTypes.String,
    E: ArrayOf(PrimitiveTypes.Float32),
    I: PrimitiveTypes.Float64,
    O: PrimitiveTypes.Float64,
}) {}
