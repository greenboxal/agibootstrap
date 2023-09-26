import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { GraphEmbedding } from "@psidb/psidb-sdk/types/psidb.indexing/GraphEmbedding";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export class NodeSearchHit extends makeSchema("psidb.indexing/NodeSearchHit", {
    indexedItem: makeSchema("", {
        chunkIndex: PrimitiveTypes.Float64,
        embeddings: GraphEmbedding,
        index: PrimitiveTypes.Float64,
        path: PrimitiveTypes.String,
    }),
    node: Node,
    score: PrimitiveTypes.Float64,
}) {}
