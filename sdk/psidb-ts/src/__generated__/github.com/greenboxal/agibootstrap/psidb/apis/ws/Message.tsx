import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { AckMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/AckMessage";
import { ErrorMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/ErrorMessage";
import { NackMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/NackMessage";
import { NotificationMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/NotificationMessage";
import { SessionMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/SessionMessage";
import { SubscribeMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/SubscribeMessage";
import { UnsubscribeMessage } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/apis/ws/UnsubscribeMessage";


export class Message extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/ws/Message", {
    ack: AckMessage,
    error: ErrorMessage,
    mid: PrimitiveTypes.Float64,
    nack: NackMessage,
    notify: NotificationMessage,
    reply_to: PrimitiveTypes.Float64,
    session: SessionMessage,
    subscribe: SubscribeMessage,
    unsubscribe: UnsubscribeMessage,
}) {}
