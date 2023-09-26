import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class GraphEmbedding extends makeSchema("psidb.indexing/GraphEmbedding", {
    depth: PrimitiveTypes.Float64,
    referenceDistance: PrimitiveTypes.Float64,
    semantic: ArrayOf(PrimitiveTypes.Float32),
    time: PrimitiveTypes.Float64,
    treeDistance: PrimitiveTypes.Float64,
}) {}
