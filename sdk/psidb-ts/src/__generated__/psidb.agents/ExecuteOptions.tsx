import { makeSchema, PrimitiveTypes, MapOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Client } from "@psidb/psidb-sdk/types/github.com/greenboxal/aip/aip-langchain/pkg/providers/openai/Client";
import { StreamedResultParser } from "@psidb/psidb-sdk/types/psidb.agents/StreamedResultParser";


export class ExecuteOptions extends makeSchema("psidb.agents/ExecuteOptions", {
    Client: Client,
    ModelOptions: makeSchema("", {
        force_function_call: PrimitiveTypes.String,
        frequency_penalty: PrimitiveTypes.Float32,
        logit_bias: MapOf(PrimitiveTypes.String, PrimitiveTypes.Integer),
        max_tokens: PrimitiveTypes.Integer,
        model: PrimitiveTypes.String,
        presence_penalty: PrimitiveTypes.Float32,
        stop: ArrayOf(PrimitiveTypes.String),
        temperature: PrimitiveTypes.Float32,
        top_p: PrimitiveTypes.Float32,
    }),
    StreamingParsers: ArrayOf(StreamedResultParser),
}) {}
