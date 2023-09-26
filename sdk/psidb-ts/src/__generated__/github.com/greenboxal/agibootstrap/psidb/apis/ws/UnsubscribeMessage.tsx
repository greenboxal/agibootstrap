import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class UnsubscribeMessage extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/ws/UnsubscribeMessage", {
    topic: PrimitiveTypes.String,
}) {}
