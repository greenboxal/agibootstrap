import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Page extends makeSchema("psidb.kb/Page", {
    body: PrimitiveTypes.String,
    title: PrimitiveTypes.String,
}) {}
