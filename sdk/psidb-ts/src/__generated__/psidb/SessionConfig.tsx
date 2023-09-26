import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { CheckpointConfig } from "@psidb/psidb-sdk/types/psidb/CheckpointConfig";
import { DataStoreConfig } from "@psidb/psidb-sdk/types/psidb/DataStoreConfig";
import { JournalConfig } from "@psidb/psidb-sdk/types/psidb/JournalConfig";
import { MountDefinition } from "@psidb/psidb-sdk/types/psidb/MountDefinition";


export class SessionConfig extends makeSchema("psidb/SessionConfig", {
    checkpoint: CheckpointConfig,
    deadline: PrimitiveTypes.String,
    graph_store: DataStoreConfig,
    journal: JournalConfig,
    keep_alive_timeout: PrimitiveTypes.Float64,
    metadata_store: DataStoreConfig,
    mount_points: ArrayOf(MountDefinition),
    parent_session_id: PrimitiveTypes.String,
    persistent: PrimitiveTypes.Boolean,
    root: PrimitiveTypes.String,
    session_id: PrimitiveTypes.String,
}) {}
