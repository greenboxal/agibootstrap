import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class File extends makeSchema("vfs/File", {
    name: PrimitiveTypes.String,
    path: PrimitiveTypes.String,
}) {}
