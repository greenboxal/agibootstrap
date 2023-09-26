import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SubscribeMessage extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/ws/SubscribeMessage", {
    depth: PrimitiveTypes.Float64,
    topic: PrimitiveTypes.String,
}) {}
