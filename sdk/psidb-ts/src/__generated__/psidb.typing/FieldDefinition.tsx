import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FieldDefinition extends makeSchema("psidb.typing/FieldDefinition", {
    name: PrimitiveTypes.String,
    type: PrimitiveTypes.String,
}) {}
