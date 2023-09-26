import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Result extends makeSchema("github.com/sashabaranov/go-openai/Result", {
    categories: makeSchema("", {
        hate: PrimitiveTypes.Boolean,
        "hate/threatening": PrimitiveTypes.Boolean,
        "self-harm": PrimitiveTypes.Boolean,
        sexual: PrimitiveTypes.Boolean,
        "sexual/minors": PrimitiveTypes.Boolean,
        violence: PrimitiveTypes.Boolean,
        "violence/graphic": PrimitiveTypes.Boolean,
    }),
    category_scores: makeSchema("", {
        hate: PrimitiveTypes.Float64,
        "hate/threatening": PrimitiveTypes.Float64,
        "self-harm": PrimitiveTypes.Float64,
        sexual: PrimitiveTypes.Float64,
        "sexual/minors": PrimitiveTypes.Float64,
        violence: PrimitiveTypes.Float64,
        "violence/graphic": PrimitiveTypes.Float64,
    }),
    flagged: PrimitiveTypes.Boolean,
}) {}
