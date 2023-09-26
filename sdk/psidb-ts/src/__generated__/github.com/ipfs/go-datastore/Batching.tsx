import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Batch } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/Batch";
import { error } from "@psidb/psidb-sdk/types//error";
import { Key } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/Key";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Query } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Query";
import { Results } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Results";


export const Batching = makeInterface({
    name: "github.com/ipfs/go-datastore/Batching",
    methods: {
        Batch: PrimitiveTypes.Func(Context)(Batch, error),
        Close: PrimitiveTypes.Func()(error),
        Delete: PrimitiveTypes.Func(Context, Key)(error),
        Get: PrimitiveTypes.Func(Context, Key)(PrimitiveTypes.Array(uint8)(error)),
        GetSize: PrimitiveTypes.Func(Context, Key)(PrimitiveTypes.Integer, error),
        Has: PrimitiveTypes.Func(Context, Key)(bool, error),
        Put: PrimitiveTypes.Func(Context, Key, PrimitiveTypes.Array(uint8))(error),
        Query: PrimitiveTypes.Func(Context, Query)(Results, error),
        Sync: PrimitiveTypes.Func(Context, Key)(error),
    },
});
