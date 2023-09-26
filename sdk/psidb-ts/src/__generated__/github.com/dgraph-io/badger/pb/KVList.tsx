import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { KV } from "@psidb/psidb-sdk/types/github.com/dgraph-io/badger/pb/KV";


export class KVList extends makeSchema("github.com/dgraph-io/badger/pb/KVList", {
    kv: ArrayOf(KV),
}) {}
