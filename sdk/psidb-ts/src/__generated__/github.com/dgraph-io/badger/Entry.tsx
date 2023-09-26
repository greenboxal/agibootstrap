import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class Entry extends makeSchema("github.com/dgraph-io/badger/Entry", {
    ExpiresAt: PrimitiveTypes.Float64,
    Key: ArrayOf(uint8),
    UserMeta: PrimitiveTypes.Float64,
    Value: ArrayOf(uint8),
}) {}
