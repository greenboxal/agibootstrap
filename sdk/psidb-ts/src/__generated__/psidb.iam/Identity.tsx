import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Identity extends makeSchema("psidb.iam/Identity", {
    username: PrimitiveTypes.String,
}) {}
