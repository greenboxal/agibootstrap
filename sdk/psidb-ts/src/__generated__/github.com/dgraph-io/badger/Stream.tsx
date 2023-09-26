import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class Stream extends makeSchema("github.com/dgraph-io/badger/Stream", {
    LogPrefix: PrimitiveTypes.String,
    NumGo: PrimitiveTypes.Float64,
    Prefix: ArrayOf(uint8),
}) {}
