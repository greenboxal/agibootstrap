import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class QueryTerm extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/psiml/QueryTerm", {
    node: Node,
    path: PrimitiveTypes.String,
    text: PrimitiveTypes.String,
}) {}
