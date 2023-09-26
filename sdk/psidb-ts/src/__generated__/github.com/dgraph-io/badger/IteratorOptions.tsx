import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class IteratorOptions extends makeSchema("github.com/dgraph-io/badger/IteratorOptions", {
    AllVersions: PrimitiveTypes.Boolean,
    InternalAccess: PrimitiveTypes.Boolean,
    PrefetchSize: PrimitiveTypes.Float64,
    PrefetchValues: PrimitiveTypes.Boolean,
    Prefix: ArrayOf(uint8),
    Reverse: PrimitiveTypes.Boolean,
}) {}
