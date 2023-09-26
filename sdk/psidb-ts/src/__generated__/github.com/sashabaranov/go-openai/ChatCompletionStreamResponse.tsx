import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { ChatCompletionStreamChoice } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ChatCompletionStreamChoice";
import { PromptAnnotation } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/PromptAnnotation";


export class ChatCompletionStreamResponse extends makeSchema("github.com/sashabaranov/go-openai/ChatCompletionStreamResponse", {
    choices: ArrayOf(ChatCompletionStreamChoice),
    created: PrimitiveTypes.Float64,
    id: PrimitiveTypes.String,
    model: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
    prompt_annotations: ArrayOf(PromptAnnotation),
}) {}
