import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class Search extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/psiml/Search", {
    from: PrimitiveTypes.String,
    limit: PrimitiveTypes.Float64,
    query: makeSchema("", {
        node: Node,
        path: PrimitiveTypes.String,
        text: PrimitiveTypes.String,
    }),
    view: PrimitiveTypes.String,
}) {}
