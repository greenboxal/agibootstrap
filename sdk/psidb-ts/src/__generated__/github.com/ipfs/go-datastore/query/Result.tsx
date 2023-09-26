import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class Result extends makeSchema("github.com/ipfs/go-datastore/query/Result", {
    Error: error,
    Expiration: PrimitiveTypes.String,
    Key: PrimitiveTypes.String,
    Size: PrimitiveTypes.Float64,
    Value: ArrayOf(uint8),
}) {}
