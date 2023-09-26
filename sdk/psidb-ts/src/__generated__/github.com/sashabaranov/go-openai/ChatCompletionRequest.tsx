import { makeSchema, PrimitiveTypes, ArrayOf, MapOf } from "@psidb/psidb-sdk/client/schema";
import { FunctionDefinition } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionDefinition";
import { ChatCompletionMessage } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ChatCompletionMessage";


export class ChatCompletionRequest extends makeSchema("github.com/sashabaranov/go-openai/ChatCompletionRequest", {
    frequency_penalty: PrimitiveTypes.Float64,
    function_call: PrimitiveTypes.Any,
    functions: ArrayOf(FunctionDefinition),
    logit_bias: MapOf(PrimitiveTypes.String, PrimitiveTypes.Integer),
    max_tokens: PrimitiveTypes.Float64,
    messages: ArrayOf(ChatCompletionMessage),
    model: PrimitiveTypes.String,
    n: PrimitiveTypes.Float64,
    presence_penalty: PrimitiveTypes.Float64,
    stop: ArrayOf(PrimitiveTypes.String),
    stream: PrimitiveTypes.Boolean,
    temperature: PrimitiveTypes.Float64,
    top_p: PrimitiveTypes.Float64,
    user: PrimitiveTypes.String,
}) {}
