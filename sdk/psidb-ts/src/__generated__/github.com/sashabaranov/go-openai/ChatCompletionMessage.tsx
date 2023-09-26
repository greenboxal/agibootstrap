import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FunctionCall } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionCall";


export class ChatCompletionMessage extends makeSchema("github.com/sashabaranov/go-openai/ChatCompletionMessage", {
    content: PrimitiveTypes.String,
    function_call: FunctionCall,
    name: PrimitiveTypes.String,
    role: PrimitiveTypes.String,
}) {}
