import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { error } from "@psidb/psidb-sdk/types//error";
import { Key } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/Key";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export const Batch = makeInterface({
    name: "github.com/ipfs/go-datastore/Batch",
    methods: {
        Commit: PrimitiveTypes.Func(Context)(error),
        Delete: PrimitiveTypes.Func(Context, Key)(error),
        Put: PrimitiveTypes.Func(Context, Key, PrimitiveTypes.Array(uint8))(error),
    },
});
