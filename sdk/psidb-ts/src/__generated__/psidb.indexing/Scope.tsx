import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Scope extends makeSchema("psidb.indexing/Scope", {
    index_name: PrimitiveTypes.String,
}) {}
