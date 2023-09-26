import { makeSchema, PrimitiveTypes, MapOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";
import { Client } from "@psidb/psidb-sdk/types/github.com/greenboxal/aip/aip-langchain/pkg/providers/openai/Client";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { ChatHistory } from "@psidb/psidb-sdk/types/psidb.agents/ChatHistory";
import { ChatLog } from "@psidb/psidb-sdk/types/psidb.agents/ChatLog";


export class ThreadContext extends makeSchema("psidb.agents/ThreadContext", {
    BaseMessage: Message,
    Client: Client,
    Ctx: Context,
    History: ChatHistory,
    Log: ChatLog,
    ModelOptions: makeSchema("", {
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
}) {}
