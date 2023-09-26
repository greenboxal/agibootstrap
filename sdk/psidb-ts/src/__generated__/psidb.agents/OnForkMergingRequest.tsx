import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Conversation } from "@psidb/psidb-sdk/types/psidb.agents/Conversation";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";


export class OnForkMergingRequest extends makeSchema("psidb.agents/OnForkMergingRequest", {
    fork: Reference(PrimitiveTypes.Pointer(Conversation)),
    merge_point: Reference(PrimitiveTypes.Pointer(Message)),
}) {}
