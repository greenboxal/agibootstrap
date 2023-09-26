import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class struct extends makeSchema("struct", {
    avg_logprob: PrimitiveTypes.Float64,
    compression_ratio: PrimitiveTypes.Float64,
    end: PrimitiveTypes.Float64,
    id: PrimitiveTypes.Float64,
    no_speech_prob: PrimitiveTypes.Float64,
    seek: PrimitiveTypes.Float64,
    start: PrimitiveTypes.Float64,
    temperature: PrimitiveTypes.Float64,
    text: PrimitiveTypes.String,
    tokens: ArrayOf(PrimitiveTypes.Integer),
    transient: PrimitiveTypes.Boolean,
}) {}
