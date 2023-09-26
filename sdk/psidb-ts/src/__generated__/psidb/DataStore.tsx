import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Batch } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/Batch";
import { error } from "@psidb/psidb-sdk/types//error";
import { DB } from "@psidb/psidb-sdk/types/github.com/dgraph-io/badger/DB";
import { Key } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/Key";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { LinkSystem } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/LinkSystem";
import { Query } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Query";
import { Results } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Results";


export const DataStore = makeInterface({
    name: "psidb/DataStore",
    methods: {
        Batch: PrimitiveTypes.Func(Context)(Batch, error),
        Close: PrimitiveTypes.Func()(error),
        DB: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(DB)),
        Delete: PrimitiveTypes.Func(Context, Key)(error),
        Get: PrimitiveTypes.Func(Context, Key)(PrimitiveTypes.Array(uint8)(error)),
        GetSize: PrimitiveTypes.Func(Context, Key)(PrimitiveTypes.Integer, error),
        Has: PrimitiveTypes.Func(Context, Key)(bool, error),
        LinkSystem: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(LinkSystem)),
        Put: PrimitiveTypes.Func(Context, Key, PrimitiveTypes.Array(uint8))(error),
        Query: PrimitiveTypes.Func(Context, Query)(Results, error),
        Sync: PrimitiveTypes.Func(Context, Key)(error),
    },
});
