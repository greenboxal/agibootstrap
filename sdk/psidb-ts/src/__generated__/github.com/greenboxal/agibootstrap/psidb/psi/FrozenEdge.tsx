import { makeSchema, MapOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/cid/Link";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class FrozenEdge extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/FrozenEdge", {
    attr: MapOf(PrimitiveTypes.String, PrimitiveTypes.Any),
    data: Link,
    from_index: PrimitiveTypes.Float64,
    from_path: PrimitiveTypes.String,
    key: PrimitiveTypes.String,
    to_index: PrimitiveTypes.Float64,
    to_link: Link,
    to_path: Path,
}) {}
