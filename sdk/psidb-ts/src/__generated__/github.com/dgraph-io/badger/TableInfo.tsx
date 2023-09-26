import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class TableInfo extends makeSchema("github.com/dgraph-io/badger/TableInfo", {
    ID: PrimitiveTypes.Float64,
    KeyCount: PrimitiveTypes.Float64,
    Left: ArrayOf(uint8),
    Level: PrimitiveTypes.Float64,
    Right: ArrayOf(uint8),
}) {}
