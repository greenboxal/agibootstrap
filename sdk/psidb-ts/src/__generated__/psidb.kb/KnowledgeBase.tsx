import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class KnowledgeBase extends makeSchema("psidb.kb/KnowledgeBase", {
    name: PrimitiveTypes.String,
}) {}
