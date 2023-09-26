import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Promise } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Promise";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class Task extends makeSchema("psidb.agents/Task", {
    completed: PrimitiveTypes.Boolean,
    error: PrimitiveTypes.Boolean,
    error_message: PrimitiveTypes.String,
    notification: makeSchema("", {
        action: PrimitiveTypes.String,
        dependencies: ArrayOf(Promise),
        interface: PrimitiveTypes.String,
        nonce: PrimitiveTypes.Float64,
        notified: PrimitiveTypes.String,
        notifier: PrimitiveTypes.String,
        observers: ArrayOf(Promise),
        params: ArrayOf(uint8),
        session_id: PrimitiveTypes.String,
        trace_id: PrimitiveTypes.String,
        trace_tags: ArrayOf(PrimitiveTypes.String),
    }),
    progress_current: PrimitiveTypes.Float64,
    progress_total: PrimitiveTypes.Float64,
    result: Path,
    uuid: PrimitiveTypes.String,
}) {}
