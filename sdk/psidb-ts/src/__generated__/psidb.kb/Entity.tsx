import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Entity extends makeSchema("psidb.kb/Entity", {
    index: PrimitiveTypes.Float64,
    title: PrimitiveTypes.String,
}) {}
