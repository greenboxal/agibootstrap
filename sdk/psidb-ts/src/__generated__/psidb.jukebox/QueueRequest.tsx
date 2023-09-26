import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class QueueRequest extends makeSchema("psidb.jukebox/QueueRequest", {
    path: PrimitiveTypes.String,
}) {}
