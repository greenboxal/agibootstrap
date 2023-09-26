import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { PackageFileManifest } from "@psidb/psidb-sdk/types/psidb.vm/PackageFileManifest";
import { ModuleManifest } from "@psidb/psidb-sdk/types/psidb.vm/ModuleManifest";


export class PackageManifest extends makeSchema("psidb.vm/PackageManifest", {
    files: ArrayOf(PackageFileManifest),
    modules: ArrayOf(ModuleManifest),
    name: PrimitiveTypes.String,
}) {}
