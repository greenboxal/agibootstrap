import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { SessionMessage } from "@psidb/psidb-sdk/types/psidb/SessionMessage";


export class SessionMessage extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/ws/SessionMessage", {
    message: SessionMessage,
    session_id: PrimitiveTypes.String,
}) {}
