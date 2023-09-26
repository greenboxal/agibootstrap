import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class KV extends makeSchema("github.com/dgraph-io/badger/pb/KV", {
    expires_at: PrimitiveTypes.Float64,
    key: ArrayOf(uint8),
    meta: ArrayOf(uint8),
    stream_done: PrimitiveTypes.Boolean,
    stream_id: PrimitiveTypes.Float64,
    user_meta: ArrayOf(uint8),
    value: ArrayOf(uint8),
    version: PrimitiveTypes.Float64,
}) {}
