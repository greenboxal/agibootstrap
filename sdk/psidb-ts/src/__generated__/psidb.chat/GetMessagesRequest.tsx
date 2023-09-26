import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";


export class GetMessagesRequest extends makeSchema("psidb.chat/GetMessagesRequest", {
    from: Reference(PrimitiveTypes.Pointer(Message)),
    skip_base_history: PrimitiveTypes.Boolean,
    to: Reference(PrimitiveTypes.Pointer(Message)),
}) {}
