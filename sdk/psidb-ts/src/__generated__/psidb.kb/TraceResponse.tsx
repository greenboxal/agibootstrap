import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class TraceResponse extends makeSchema("psidb.kb/TraceResponse", {
    trace: ArrayOf(Path),
}) {}
