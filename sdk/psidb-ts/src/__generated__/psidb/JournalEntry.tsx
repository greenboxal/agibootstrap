import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Confirmation } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Confirmation";
import { SerializedEdge } from "@psidb/psidb-sdk/types/psidb/SerializedEdge";
import { SerializedNode } from "@psidb/psidb-sdk/types/psidb/SerializedNode";
import { Notification } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Notification";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";
import { Promise } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Promise";


export class JournalEntry extends makeSchema("psidb/JournalEntry", {
    confirmation: Confirmation,
    edge: SerializedEdge,
    inode: PrimitiveTypes.Float64,
    node: SerializedNode,
    notification: Notification,
    op: PrimitiveTypes.Float64,
    path: Path,
    promises: ArrayOf(Promise),
    rid: PrimitiveTypes.Float64,
    ts: PrimitiveTypes.Float64,
    xid: PrimitiveTypes.Float64,
}) {}
