import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class IndexNodeRequest extends makeSchema("psidb.indexing/IndexNodeRequest", {
    chunkIndex: PrimitiveTypes.Float64,
    embeddings: makeSchema("", {
        depth: PrimitiveTypes.Float64,
        referenceDistance: PrimitiveTypes.Float64,
        semantic: ArrayOf(PrimitiveTypes.Float32),
        time: PrimitiveTypes.Float64,
        treeDistance: PrimitiveTypes.Float64,
    }),
    path: PrimitiveTypes.String,
}) {}
