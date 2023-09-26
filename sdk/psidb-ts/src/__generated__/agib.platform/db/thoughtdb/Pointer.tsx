import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Pointer extends makeSchema("agib.platform/db/thoughtdb/Pointer", {
    clock: PrimitiveTypes.Float64,
    level: PrimitiveTypes.Float64,
    parent: PrimitiveTypes.String,
    previous: PrimitiveTypes.String,
    timestamp: PrimitiveTypes.String,
}) {}
