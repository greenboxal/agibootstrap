import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { MountTarget } from "@psidb/psidb-sdk/types/psidb/MountTarget";


export class MountDefinition extends makeSchema("psidb/MountDefinition", {
    name: PrimitiveTypes.String,
    path: PrimitiveTypes.String,
    target: MountTarget,
}) {}
