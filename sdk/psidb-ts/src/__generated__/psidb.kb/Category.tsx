import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Category extends makeSchema("psidb.kb/Category", {
    description: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
    slug: PrimitiveTypes.String,
}) {}
