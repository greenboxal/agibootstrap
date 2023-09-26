import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FunctionCall } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionCall";


export class ChatCompletionChoice extends makeSchema("github.com/sashabaranov/go-openai/ChatCompletionChoice", {
    finish_reason: PrimitiveTypes.String,
    index: PrimitiveTypes.Float64,
    message: makeSchema("", {
        content: PrimitiveTypes.String,
        function_call: FunctionCall,
        name: PrimitiveTypes.String,
        role: PrimitiveTypes.String,
    }),
}) {}
