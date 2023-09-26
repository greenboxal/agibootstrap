import { makeSchema, PrimitiveTypes, MapOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Reference } from "@psidb/psidb-sdk/types/stdlib/Reference";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";
import { PromptToolSelection } from "@psidb/psidb-sdk/types/psidb.agents/PromptToolSelection";


export class OnMessageSideEffectRequest extends makeSchema("psidb.agents/OnMessageSideEffectRequest", {
    message: Reference(PrimitiveTypes.Pointer(Message)),
    options: makeSchema("", {
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
    tool_selection: PromptToolSelection,
}) {}
