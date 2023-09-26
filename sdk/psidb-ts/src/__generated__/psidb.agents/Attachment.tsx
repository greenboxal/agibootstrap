import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Attachment extends makeSchema("psidb.agents/Attachment", {
    name: PrimitiveTypes.String,
}) {}
