import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { chan } from "@psidb/psidb-sdk/types//chan";
import { Result } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Result";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Process } from "@psidb/psidb-sdk/types/github.com/jbenet/goprocess/Process";
import { Query } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Query";
import { Entry } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Entry";


export const Results = makeInterface({
    name: "github.com/ipfs/go-datastore/query/Results",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Next: PrimitiveTypes.Func()(chan),
        NextSync: PrimitiveTypes.Func()(Result, bool),
        Process: PrimitiveTypes.Func()(Process),
        Query: PrimitiveTypes.Func()(Query),
        Rest: PrimitiveTypes.Func()(PrimitiveTypes.Array(Entry)(error)),
    },
});
