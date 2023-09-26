import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { GraphEmbedding } from "@psidb/psidb-sdk/types/psidb.indexing/GraphEmbedding";


export class BasicSearchHit extends makeSchema("psidb.indexing/BasicSearchHit", {
    indexedItem: makeSchema("", {
        chunkIndex: PrimitiveTypes.Float64,
        embeddings: GraphEmbedding,
        index: PrimitiveTypes.Float64,
        path: PrimitiveTypes.String,
    }),
    score: PrimitiveTypes.Float64,
}) {}
