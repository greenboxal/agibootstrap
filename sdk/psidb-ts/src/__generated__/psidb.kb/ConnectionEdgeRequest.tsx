import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ConnectionEdgeRequest extends makeSchema("psidb.kb/ConnectionEdgeRequest", {
    expand: PrimitiveTypes.Boolean,
    frontier: PrimitiveTypes.String,
    observers: makeSchema("", {
        c: PrimitiveTypes.Float64,
        nonce: PrimitiveTypes.Float64,
        xid: PrimitiveTypes.Float64,
    }),
    to: PrimitiveTypes.String,
}) {}
