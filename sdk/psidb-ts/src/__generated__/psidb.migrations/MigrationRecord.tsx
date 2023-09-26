import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class MigrationRecord extends makeSchema("psidb.migrations/MigrationRecord", {
    name: PrimitiveTypes.String,
    timestamp: PrimitiveTypes.String,
}) {}
