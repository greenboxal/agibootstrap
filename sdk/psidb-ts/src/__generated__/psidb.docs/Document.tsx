import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Document extends makeSchema("psidb.docs/Document", {
    content: PrimitiveTypes.String,
    observations: PrimitiveTypes.String,
    title: PrimitiveTypes.String,
    uuid: PrimitiveTypes.String,
}) {}
