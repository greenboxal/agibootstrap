import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Notification extends makeSchema("psidb.pubsub/Notification", {
    path: PrimitiveTypes.String,
    ts: PrimitiveTypes.Float64,
}) {}
