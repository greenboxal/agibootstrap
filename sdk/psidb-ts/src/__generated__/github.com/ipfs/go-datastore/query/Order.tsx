import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Entry } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Entry";


export const Order = makeInterface({
    name: "github.com/ipfs/go-datastore/query/Order",
    methods: {
        Compare: PrimitiveTypes.Func(Entry, Entry)(PrimitiveTypes.Integer),
    },
});
