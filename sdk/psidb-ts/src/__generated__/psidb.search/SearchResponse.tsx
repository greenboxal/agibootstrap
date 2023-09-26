import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { NodeSearchHit } from "@psidb/psidb-sdk/types/psidb.indexing/NodeSearchHit";


export class SearchResponse extends makeSchema("psidb.search/SearchResponse", {
    results: ArrayOf(NodeSearchHit),
}) {}
