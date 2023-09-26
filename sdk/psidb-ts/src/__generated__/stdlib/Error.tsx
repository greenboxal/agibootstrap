import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Error extends makeSchema("stdlib/Error", {
    message: PrimitiveTypes.String,
}) {}
