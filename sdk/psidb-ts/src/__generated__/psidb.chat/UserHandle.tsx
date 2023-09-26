import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class UserHandle extends makeSchema("psidb.chat/UserHandle", {
    id: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
    role: PrimitiveTypes.String,
}) {}
