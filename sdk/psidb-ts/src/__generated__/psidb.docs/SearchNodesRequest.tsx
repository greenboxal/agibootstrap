import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class SearchNodesRequest extends makeSchema("psidb.docs/SearchNodesRequest", {
    limit: PrimitiveTypes.Float64,
    query_node: Reference(Node),
    query_prompt: PrimitiveTypes.String,
}) {}
