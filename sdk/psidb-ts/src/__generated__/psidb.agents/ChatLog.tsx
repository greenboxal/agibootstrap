import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";
import { PromptResponseChoice } from "@psidb/psidb-sdk/types/psidb.agents/PromptResponseChoice";
import { error } from "@psidb/psidb-sdk/types//error";
import { ModelOptions } from "@psidb/psidb-sdk/types/gpt/ModelOptions";

const _F = {} as any

export const ChatLog = makeInterface({
    name: "psidb.agents/ChatLog",
    methods: {
        AcceptChoice: PrimitiveTypes.Func(Context, PrimitiveTypes.Pointer(Message)(PromptResponseChoice))(error),
        AcceptMessage: PrimitiveTypes.Func(Context, PrimitiveTypes.Pointer(Message))(error),
        ForkAsChatLog: PrimitiveTypes.Func(Context, PrimitiveTypes.Pointer(Message)(ModelOptions))(_F["ChatLog"], error),
        MessageIterator: PrimitiveTypes.Func(Context)(Iterator(PrimitiveTypes.Pointer(Message))(error)),
    },
});
_F["ChatLog"] = ChatLog;
