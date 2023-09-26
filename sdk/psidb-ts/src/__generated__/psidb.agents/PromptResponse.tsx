import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { PromptResponseChoice } from "@psidb/psidb-sdk/types/psidb.agents/PromptResponseChoice";
import { Trace } from "@psidb/psidb-sdk/types/gpt/Trace";


export class PromptResponse extends makeSchema("psidb.agents/PromptResponse", {
    choices: ArrayOf(PromptResponseChoice),
    raw: Trace,
}) {}
