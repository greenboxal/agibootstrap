import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class Entry extends makeSchema("github.com/ipfs/go-datastore/query/Entry", {
    Expiration: PrimitiveTypes.String,
    Key: PrimitiveTypes.String,
    Size: PrimitiveTypes.Float64,
    Value: ArrayOf(uint8),
}) {}
