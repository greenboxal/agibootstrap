import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Index extends makeSchema("psidb.docs/Index", {
    counter: PrimitiveTypes.Float64,
    name: PrimitiveTypes.String,
    uuid: PrimitiveTypes.String,
}) {}
