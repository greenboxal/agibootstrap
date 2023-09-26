import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/cid/Link";


export class SerializedEdge extends makeSchema("psidb/SerializedEdge", {
    data: ArrayOf(uint8),
    flags: PrimitiveTypes.Float64,
    index: PrimitiveTypes.Float64,
    key: PrimitiveTypes.String,
    toIndex: PrimitiveTypes.Float64,
    toLink: Link,
    toPath: PrimitiveTypes.String,
    version: PrimitiveTypes.Float64,
    xmax: PrimitiveTypes.Float64,
    xmin: PrimitiveTypes.Float64,
}) {}
