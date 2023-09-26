import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/cid/Link";


export class SerializedNode extends makeSchema("psidb/SerializedNode", {
    data: ArrayOf(uint8),
    flags: PrimitiveTypes.Float64,
    index: PrimitiveTypes.Float64,
    link: Link,
    parent: PrimitiveTypes.Float64,
    path: PrimitiveTypes.String,
    type: PrimitiveTypes.String,
    version: PrimitiveTypes.Float64,
    xmax: PrimitiveTypes.Float64,
    xmin: PrimitiveTypes.Float64,
}) {}
