import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class CommHandle extends makeSchema("agib.platform/db/thoughtdb/CommHandle", {
    ID: PrimitiveTypes.String,
    Name: PrimitiveTypes.String,
    Role: PrimitiveTypes.String,
}) {}
