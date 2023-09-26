import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";
import { PromptToolSelection } from "@psidb/psidb-sdk/types/psidb.agents/PromptToolSelection";


export class PromptResponseChoice extends makeSchema("psidb.agents/PromptResponseChoice", {
    index: PrimitiveTypes.Float64,
    message: Message,
    reason: PrimitiveTypes.String,
    tool: PromptToolSelection,
}) {}
