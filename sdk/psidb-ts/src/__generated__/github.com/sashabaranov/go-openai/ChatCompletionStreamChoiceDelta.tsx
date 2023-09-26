import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FunctionCall } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionCall";


export class ChatCompletionStreamChoiceDelta extends makeSchema("github.com/sashabaranov/go-openai/ChatCompletionStreamChoiceDelta", {
    content: PrimitiveTypes.String,
    function_call: FunctionCall,
    role: PrimitiveTypes.String,
}) {}
