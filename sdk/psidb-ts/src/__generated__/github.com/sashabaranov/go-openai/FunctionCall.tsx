import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FunctionCall extends makeSchema("github.com/sashabaranov/go-openai/FunctionCall", {
    arguments: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
}) {}
