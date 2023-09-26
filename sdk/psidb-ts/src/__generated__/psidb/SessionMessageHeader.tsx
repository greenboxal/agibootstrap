import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SessionMessageHeader extends makeSchema("psidb/SessionMessageHeader", {
    message_id: PrimitiveTypes.Float64,
    reply_to_id: PrimitiveTypes.Float64,
    session_id: PrimitiveTypes.String,
}) {}
