import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PromptAnnotation extends makeSchema("github.com/sashabaranov/go-openai/PromptAnnotation", {
    content_filter_results: makeSchema("", {
        hate: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
        self_harm: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
        sexual: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
        violence: makeSchema("", {
            filtered: PrimitiveTypes.Boolean,
            severity: PrimitiveTypes.String,
        }),
    }),
    prompt_index: PrimitiveTypes.Float64,
}) {}
