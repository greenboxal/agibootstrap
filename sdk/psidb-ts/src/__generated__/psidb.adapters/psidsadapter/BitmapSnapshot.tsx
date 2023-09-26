import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class BitmapSnapshot extends makeSchema("psidb.adapters/psidsadapter/BitmapSnapshot", {
    freeList: ArrayOf(uint8),
    lastID: PrimitiveTypes.Float64,
    usedList: ArrayOf(uint8),
}) {}
