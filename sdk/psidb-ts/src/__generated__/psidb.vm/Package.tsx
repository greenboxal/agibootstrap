import { makeSchema, MapOf, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { PackageFileManifest } from "@psidb/psidb-sdk/types/psidb.vm/PackageFileManifest";
import { ModuleManifest } from "@psidb/psidb-sdk/types/psidb.vm/ModuleManifest";


export class Package extends makeSchema("psidb.vm/Package", {
    files: MapOf(PrimitiveTypes.String, PrimitiveTypes.String),
    manifest: makeSchema("", {
        files: ArrayOf(PackageFileManifest),
        modules: ArrayOf(ModuleManifest),
        name: PrimitiveTypes.String,
    }),
    name: PrimitiveTypes.String,
    registered: PrimitiveTypes.Boolean,
}) {}
