import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class Document extends makeSchema("psidb.kb/Document", {
    body: PrimitiveTypes.String,
    categories: ArrayOf(PrimitiveTypes.String),
    description: PrimitiveTypes.String,
    has_content: PrimitiveTypes.Boolean,
    has_summary: PrimitiveTypes.Boolean,
    related_topics: ArrayOf(PrimitiveTypes.String),
    root: PrimitiveTypes.String,
    slug: PrimitiveTypes.String,
    summary: PrimitiveTypes.String,
    title: PrimitiveTypes.String,
}) {}
