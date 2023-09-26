import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PackageFileManifest extends makeSchema("psidb.vm/PackageFileManifest", {
    hash: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
}) {}
