import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { FunctionCall } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionCall";


export class StreamingTraceChunk extends makeSchema("gpt/StreamingTraceChunk", {
    choice_index: PrimitiveTypes.Float64,
    content: PrimitiveTypes.String,
    finish_reason: PrimitiveTypes.String,
    function_call: FunctionCall,
    role: PrimitiveTypes.String,
    tags: ArrayOf(PrimitiveTypes.String),
    trace_id: PrimitiveTypes.String,
}) {}
