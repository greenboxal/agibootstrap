import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { ActionDefinition } from "@psidb/psidb-sdk/types/psidb.typing/ActionDefinition";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class InterfaceDefinition extends makeSchema("psidb.typing/InterfaceDefinition", {
    actions: ArrayOf(ActionDefinition),
    description: PrimitiveTypes.String,
    module: Path,
    name: PrimitiveTypes.String,
}) {}
