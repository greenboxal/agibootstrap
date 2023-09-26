import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Type } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Type";


export class NodeActionDefinition extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/NodeActionDefinition", {
    description: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
    request_type: Type,
    response_type: Type,
}) {}
