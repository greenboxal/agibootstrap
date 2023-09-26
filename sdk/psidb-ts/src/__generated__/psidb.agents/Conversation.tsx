import { makeSchema, PrimitiveTypes, MapOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";

const _F = {} as any

export class Conversation extends makeSchema("psidb.agents/Conversation", {
    base_conversation: Reference(PrimitiveTypes.Pointer(_F["Conversation"])),
    base_message: Reference(PrimitiveTypes.Pointer(Message)),
    base_options: makeSchema("", {
        force_function_call: PrimitiveTypes.String,
        frequency_penalty: PrimitiveTypes.Float32,
        logit_bias: MapOf(PrimitiveTypes.String, PrimitiveTypes.Integer),
        max_tokens: PrimitiveTypes.Integer,
        model: PrimitiveTypes.String,
        presence_penalty: PrimitiveTypes.Float32,
        stop: ArrayOf(PrimitiveTypes.String),
        temperature: PrimitiveTypes.Float32,
        top_p: PrimitiveTypes.Float32,
    }),
    is_merged: PrimitiveTypes.Boolean,
    is_title_temporary: PrimitiveTypes.Boolean,
    name: PrimitiveTypes.String,
    title: PrimitiveTypes.String,
    trace_tags: ArrayOf(PrimitiveTypes.String),
}) {}
_F["Conversation"] = Conversation;
