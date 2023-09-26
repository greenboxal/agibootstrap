import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ResultCategoryScores extends makeSchema("github.com/sashabaranov/go-openai/ResultCategoryScores", {
    hate: PrimitiveTypes.Float64,
    "hate/threatening": PrimitiveTypes.Float64,
    "self-harm": PrimitiveTypes.Float64,
    sexual: PrimitiveTypes.Float64,
    "sexual/minors": PrimitiveTypes.Float64,
    violence: PrimitiveTypes.Float64,
    "violence/graphic": PrimitiveTypes.Float64,
}) {}
