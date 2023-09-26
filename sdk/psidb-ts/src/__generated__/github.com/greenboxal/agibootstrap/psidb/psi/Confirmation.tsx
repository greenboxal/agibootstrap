import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Confirmation extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/Confirmation", {
    nonce: PrimitiveTypes.Float64,
    ok: PrimitiveTypes.Boolean,
    rid: PrimitiveTypes.Float64,
    xid: PrimitiveTypes.Float64,
}) {}
