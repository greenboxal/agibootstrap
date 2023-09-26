import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Filter } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Filter";
import { Order } from "@psidb/psidb-sdk/types/github.com/ipfs/go-datastore/query/Order";


export class Query extends makeSchema("github.com/ipfs/go-datastore/query/Query", {
    Filters: ArrayOf(Filter),
    KeysOnly: PrimitiveTypes.Boolean,
    Limit: PrimitiveTypes.Float64,
    Offset: PrimitiveTypes.Float64,
    Orders: ArrayOf(Order),
    Prefix: PrimitiveTypes.String,
    ReturnExpirations: PrimitiveTypes.Boolean,
    ReturnsSizes: PrimitiveTypes.Boolean,
}) {}
