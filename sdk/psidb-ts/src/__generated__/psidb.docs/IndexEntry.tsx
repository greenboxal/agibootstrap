import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class IndexEntry extends makeSchema("psidb.docs/IndexEntry", {
    K: PrimitiveTypes.String,
    Q: PrimitiveTypes.String,
}) {}
