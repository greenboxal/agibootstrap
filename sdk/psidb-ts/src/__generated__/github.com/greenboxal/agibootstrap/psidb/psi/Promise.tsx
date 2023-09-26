import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Promise extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/Promise", {
    c: PrimitiveTypes.Float64,
    nonce: PrimitiveTypes.Float64,
    xid: PrimitiveTypes.Float64,
}) {}
