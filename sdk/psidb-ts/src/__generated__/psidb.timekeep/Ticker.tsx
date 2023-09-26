import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Ticker extends makeSchema("psidb.timekeep/Ticker", {
    name: PrimitiveTypes.String,
    stop_at: PrimitiveTypes.Float64,
}) {}
