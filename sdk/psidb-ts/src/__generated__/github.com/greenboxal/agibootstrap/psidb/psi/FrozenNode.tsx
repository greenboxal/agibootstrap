import { makeSchema, MapOf, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { EdgeKey } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeKey";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/cid/Link";


export class FrozenNode extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/FrozenNode", {
    attr: MapOf(PrimitiveTypes.String, PrimitiveTypes.String),
    children: ArrayOf(EdgeKey),
    edges: ArrayOf(Link),
    index: PrimitiveTypes.Float64,
    link: Link,
    path: PrimitiveTypes.String,
    type: PrimitiveTypes.String,
    version: PrimitiveTypes.Float64,
}) {}
