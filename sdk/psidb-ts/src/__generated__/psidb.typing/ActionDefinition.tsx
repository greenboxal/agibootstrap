import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class ActionDefinition extends makeSchema("psidb.typing/ActionDefinition", {
    bound_function: PrimitiveTypes.String,
    description: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
    request_type: Path,
    response_type: Path,
}) {}
