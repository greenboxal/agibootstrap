import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";


export class UpdateTitleRequest extends makeSchema("psidb.agents/UpdateTitleRequest", {
    last_message: Reference(PrimitiveTypes.Pointer(Message)),
}) {}
