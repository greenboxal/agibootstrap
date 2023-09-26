import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Schema } from "@psidb/psidb-sdk/types/github.com/invopop/jsonschema/Schema";


export class ModuleManifest extends makeSchema("psidb.vm/ModuleManifest", {
    entrypoint: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
    schema: Schema,
}) {}
