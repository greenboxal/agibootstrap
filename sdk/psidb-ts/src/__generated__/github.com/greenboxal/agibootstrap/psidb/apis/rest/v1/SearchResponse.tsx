import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { NodeSearchHit } from "@psidb/psidb-sdk/types/psidb.indexing/NodeSearchHit";


export class SearchResponse extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/rest/v1/SearchResponse", {
    results: ArrayOf(NodeSearchHit),
}) {}
