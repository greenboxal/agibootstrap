import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ResultCategories extends makeSchema("github.com/sashabaranov/go-openai/ResultCategories", {
    hate: PrimitiveTypes.Boolean,
    "hate/threatening": PrimitiveTypes.Boolean,
    "self-harm": PrimitiveTypes.Boolean,
    sexual: PrimitiveTypes.Boolean,
    "sexual/minors": PrimitiveTypes.Boolean,
    violence: PrimitiveTypes.Boolean,
    "violence/graphic": PrimitiveTypes.Boolean,
}) {}
