import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FunctionDefinition } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FunctionDefinition";


export const PromptBuilderTool = makeInterface({
    name: "psidb.agents/PromptBuilderTool",
    methods: {
        ToolDefinition: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(FunctionDefinition)),
        ToolName: PrimitiveTypes.Func()(PrimitiveTypes.String),
    },
});
