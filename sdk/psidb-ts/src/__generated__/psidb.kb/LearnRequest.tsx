import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class LearnRequest extends makeSchema("psidb.kb/LearnRequest", {
    current_depth: PrimitiveTypes.Float64,
    feedback: PrimitiveTypes.String,
    max_depth: PrimitiveTypes.Float64,
    observers: makeSchema("", {
        c: PrimitiveTypes.Float64,
        nonce: PrimitiveTypes.Float64,
        xid: PrimitiveTypes.Float64,
    }),
    references: ArrayOf(Path),
}) {}
