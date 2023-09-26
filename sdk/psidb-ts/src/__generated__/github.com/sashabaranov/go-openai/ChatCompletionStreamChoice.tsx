import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FunctionCall } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionCall";


export class ChatCompletionStreamChoice extends makeSchema("github.com/sashabaranov/go-openai/ChatCompletionStreamChoice", {
    content_filter_results: makeSchema("", {
        hate: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
        self_harm: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
        sexual: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
        violence: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
    }),
    delta: makeSchema("", {
        content: PrimitiveTypes.String,
        function_call: FunctionCall,
        role: PrimitiveTypes.String,
    }),
    finish_reason: PrimitiveTypes.String,
    index: PrimitiveTypes.Float64,
}) {}
