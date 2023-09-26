import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Module extends makeSchema("psidb.vm/Module", {
    name: PrimitiveTypes.String,
    source: PrimitiveTypes.String,
    source_file: PrimitiveTypes.String,
}) {}
