import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class TraceRequest extends makeSchema("psidb.kb/TraceRequest", {
    dispatch: PrimitiveTypes.Boolean,
    from: PrimitiveTypes.String,
    to: PrimitiveTypes.String,
}) {}
