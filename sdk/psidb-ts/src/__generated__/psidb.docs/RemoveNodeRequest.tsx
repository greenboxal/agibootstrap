import { makeSchema } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class RemoveNodeRequest extends makeSchema("psidb.docs/RemoveNodeRequest", {
    node: Reference(Node),
}) {}
