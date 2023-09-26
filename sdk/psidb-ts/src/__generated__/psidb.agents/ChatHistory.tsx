import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";
import { ModelOptions } from "@psidb/psidb-sdk/types/gpt/ModelOptions";
import { ChatLog } from "@psidb/psidb-sdk/types/psidb.agents/ChatLog";
import { error } from "@psidb/psidb-sdk/types//error";


export const ChatHistory = makeInterface({
    name: "psidb.agents/ChatHistory",
    methods: {
        ForkAsChatLog: PrimitiveTypes.Func(Context, PrimitiveTypes.Pointer(Message)(ModelOptions))(ChatLog, error),
        MessageIterator: PrimitiveTypes.Func(Context)(Iterator(PrimitiveTypes.Pointer(Message))(error)),
    },
});
