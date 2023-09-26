import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class NodeTypeDefinition extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/NodeTypeDefinition", {
    class: PrimitiveTypes.String,
    is_runtime_only: PrimitiveTypes.Boolean,
    is_stub: PrimitiveTypes.Boolean,
    name: PrimitiveTypes.String,
}) {}
