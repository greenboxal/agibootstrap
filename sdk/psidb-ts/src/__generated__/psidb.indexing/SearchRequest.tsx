import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { GraphOperations } from "@psidb/psidb-sdk/types/psidb/GraphOperations";


export class SearchRequest extends makeSchema("psidb.indexing/SearchRequest", {
    Graph: GraphOperations,
    Limit: PrimitiveTypes.Float64,
    Query: makeSchema("", {
        depth: PrimitiveTypes.Float64,
        referenceDistance: PrimitiveTypes.Float64,
        semantic: ArrayOf(PrimitiveTypes.Float32),
        time: PrimitiveTypes.Float64,
        treeDistance: PrimitiveTypes.Float64,
    }),
    ReturnEmbeddings: PrimitiveTypes.Boolean,
    ReturnNode: PrimitiveTypes.Boolean,
}) {}
