import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Engine } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/Engine";


export class EnginesList extends makeSchema("github.com/sashabaranov/go-openai/EnginesList", {
    data: ArrayOf(Engine),
}) {}
