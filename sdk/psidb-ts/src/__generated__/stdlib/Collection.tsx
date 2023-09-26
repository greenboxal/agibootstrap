import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Collection extends makeSchema("stdlib/Collection", {
    name: PrimitiveTypes.String,
}) {}
