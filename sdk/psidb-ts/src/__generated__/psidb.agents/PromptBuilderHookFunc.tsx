import { Type, defineFunction, SequenceOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { PromptBuilder } from "@psidb/psidb-sdk/types/psidb.agents/PromptBuilder";
import { ChatCompletionRequest } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/ChatCompletionRequest";


export function PromptBuilderHookFunc<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return defineFunction(SequenceOf(Context, PromptBuilder, ChatCompletionRequest))(ArrayOf())
}