import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { ChatCompletionChoice } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ChatCompletionChoice";
import { error } from "@psidb/psidb-sdk/types//error";
import { ChatCompletionMessage } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ChatCompletionMessage";


export class SessionMessageGPTrace extends makeSchema("gpt/SessionMessageGPTrace", {
    trace: makeSchema("", {
        choices: ArrayOf(ChatCompletionChoice),
        done: PrimitiveTypes.Boolean,
        error: error,
        messages: ArrayOf(ChatCompletionMessage),
        tags: ArrayOf(PrimitiveTypes.String),
        trace_id: PrimitiveTypes.String,
    }),
}) {}
