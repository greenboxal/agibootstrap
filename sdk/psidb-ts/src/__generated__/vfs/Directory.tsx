import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Directory extends makeSchema("vfs/Directory", {
    name: PrimitiveTypes.String,
    path: PrimitiveTypes.String,
}) {}
