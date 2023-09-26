import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FieldDefinition } from "@psidb/psidb-sdk/types/psidb.typing/FieldDefinition";
import { InterfaceDefinition } from "@psidb/psidb-sdk/types/psidb.typing/InterfaceDefinition";


export class Type extends makeSchema("psidb.typing/Type", {
    fields: ArrayOf(FieldDefinition),
    full_name: PrimitiveTypes.String,
    interfaces: ArrayOf(InterfaceDefinition),
    name: PrimitiveTypes.String,
    primitive_kind: PrimitiveTypes.Float64,
}) {}
