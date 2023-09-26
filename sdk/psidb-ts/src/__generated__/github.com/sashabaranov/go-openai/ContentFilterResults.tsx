import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ContentFilterResults extends makeSchema("github.com/sashabaranov/go-openai/ContentFilterResults", {
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
}) {}
