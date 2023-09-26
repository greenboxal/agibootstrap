import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Entry } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Entry";
import { bool } from "@psidb/psidb-sdk/types//bool";


export const Filter = makeInterface({
    name: "github.com/ipfs/go-datastore/query/Filter",
    methods: {
        Filter: PrimitiveTypes.Func(Entry)(bool),
    },
});
