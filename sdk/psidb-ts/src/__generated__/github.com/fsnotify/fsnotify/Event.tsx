import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Event extends makeSchema("github.com/fsnotify/fsnotify/Event", {
    Name: PrimitiveTypes.String,
    Op: PrimitiveTypes.Float64,
}) {}
