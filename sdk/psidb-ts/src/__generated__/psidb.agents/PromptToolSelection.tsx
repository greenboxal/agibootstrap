import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class PromptToolSelection extends makeSchema("psidb.agents/PromptToolSelection", {
    arguments: PrimitiveTypes.String,
    focus: Reference(Node),
    name: PrimitiveTypes.String,
}) {}
