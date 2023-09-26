import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class DequeueRequest extends makeSchema("psidb.jukebox/DequeueRequest", {
    path: PrimitiveTypes.String,
}) {}
