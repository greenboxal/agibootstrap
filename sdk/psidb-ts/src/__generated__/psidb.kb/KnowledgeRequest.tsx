import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Document } from "@psidb/psidb-sdk/types/psidb.kb/Document";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class KnowledgeRequest extends makeSchema("psidb.kb/KnowledgeRequest", {
    back_link_to: Reference(PrimitiveTypes.Pointer(Document)),
    current_depth: PrimitiveTypes.Float64,
    description: PrimitiveTypes.String,
    max_depth: PrimitiveTypes.Float64,
    observer: makeSchema("", {
        c: PrimitiveTypes.Float64,
        nonce: PrimitiveTypes.Float64,
        xid: PrimitiveTypes.Float64,
    }),
    references: ArrayOf(Path),
    title: PrimitiveTypes.String,
}) {}
