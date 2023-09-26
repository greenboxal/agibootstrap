import { makeSchema, PrimitiveTypes, MapOf, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class ModelOptions extends makeSchema("gpt/ModelOptions", {
    force_function_call: PrimitiveTypes.String,
    frequency_penalty: PrimitiveTypes.Float32,
    logit_bias: MapOf(PrimitiveTypes.String, PrimitiveTypes.Integer),
    max_tokens: PrimitiveTypes.Integer,
    model: PrimitiveTypes.String,
    presence_penalty: PrimitiveTypes.Float32,
    stop: ArrayOf(PrimitiveTypes.String),
    temperature: PrimitiveTypes.Float32,
    top_p: PrimitiveTypes.Float32,
}) {}
