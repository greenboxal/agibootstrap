import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Tick extends makeSchema("psidb.timekeep/Tick", {
    t: PrimitiveTypes.Float64,
    x: PrimitiveTypes.Float64,
}) {}
