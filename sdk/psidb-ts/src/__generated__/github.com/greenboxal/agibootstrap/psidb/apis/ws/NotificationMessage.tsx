import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class NotificationMessage extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/ws/NotificationMessage", {
    notification: makeSchema("", {
        path: PrimitiveTypes.String,
        ts: PrimitiveTypes.Float64,
    }),
}) {}
