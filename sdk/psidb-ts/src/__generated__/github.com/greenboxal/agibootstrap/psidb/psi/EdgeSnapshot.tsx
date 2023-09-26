import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FrozenEdge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/FrozenEdge";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Link";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class EdgeSnapshot extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/EdgeSnapshot", {
    Frozen: FrozenEdge,
    Index: PrimitiveTypes.Float64,
    Link: Link,
    Node: Node,
}) {}
