import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { IndexEntryChunk } from "@psidb/psidb-sdk/types/psidb.docs/IndexEntryChunk";
import { IndexEntry } from "@psidb/psidb-sdk/types/psidb.docs/IndexEntry";


export class SearchNodeHit extends makeSchema("psidb.docs/SearchNodeHit", {
    chunk: IndexEntryChunk,
    entry: IndexEntry,
    score: PrimitiveTypes.Float64,
}) {}
