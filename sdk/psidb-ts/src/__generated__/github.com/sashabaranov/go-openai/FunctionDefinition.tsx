import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FunctionDefinition extends makeSchema("github.com/sashabaranov/go-openai/FunctionDefinition", {
    description: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
    parameters: PrimitiveTypes.Any,
}) {}
