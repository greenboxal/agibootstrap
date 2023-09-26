import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { PromptResponseChoice } from "@psidb/psidb-sdk/types/psidb.agents/PromptResponseChoice";
import { ChatCompletionStreamChoice } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ChatCompletionStreamChoice";


export const StreamedResultParser = makeInterface({
    name: "psidb.agents/StreamedResultParser",
    methods: {
        Error: PrimitiveTypes.Func()(error),
        ParseChoice: PrimitiveTypes.Func(Context, PromptResponseChoice)(error),
        ParseChoiceStreamed: PrimitiveTypes.Func(Context, ChatCompletionStreamChoice)(error),
    },
});
