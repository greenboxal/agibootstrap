import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Prefix extends makeSchema("github.com/ipfs/go-cid/Prefix", {
    Codec: PrimitiveTypes.Float64,
    MhLength: PrimitiveTypes.Float64,
    MhType: PrimitiveTypes.Float64,
    Version: PrimitiveTypes.Float64,
}) {}
