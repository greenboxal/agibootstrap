import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Pointer } from "@psidb/psidb-sdk/types/agib.platform/db/thoughtdb/Pointer";
import { CommHandle } from "@psidb/psidb-sdk/types/agib.platform/db/thoughtdb/CommHandle";


export class Thought extends makeSchema("agib.platform/db/thoughtdb/Thought", {
    From: makeSchema("", {
        ID: PrimitiveTypes.String,
        Name: PrimitiveTypes.String,
        Role: PrimitiveTypes.String,
    }),
    Parents: ArrayOf(Pointer),
    Pointer: makeSchema("", {
        clock: PrimitiveTypes.Float64,
        level: PrimitiveTypes.Float64,
        parent: PrimitiveTypes.String,
        previous: PrimitiveTypes.String,
        timestamp: PrimitiveTypes.String,
    }),
    ReplyTo: CommHandle,
    Text: PrimitiveTypes.String,
}) {}
