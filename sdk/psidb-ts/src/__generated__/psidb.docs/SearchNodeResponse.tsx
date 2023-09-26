import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { SearchNodeHit } from "@psidb/psidb-sdk/types/psidb.docs/SearchNodeHit";


export class SearchNodeResponse extends makeSchema("psidb.docs/SearchNodeResponse", {
    hits: ArrayOf(SearchNodeHit),
}) {}
