import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { GraphEmbedding } from "@psidb/psidb-sdk/types/psidb.indexing/GraphEmbedding";


export class IndexedItem extends makeSchema("psidb.indexing/IndexedItem", {
    chunkIndex: PrimitiveTypes.Float64,
    embeddings: GraphEmbedding,
    index: PrimitiveTypes.Float64,
    path: PrimitiveTypes.String,
}) {}
