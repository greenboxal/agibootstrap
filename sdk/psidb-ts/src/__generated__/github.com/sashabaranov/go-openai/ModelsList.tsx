import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Model } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/Model";


export class ModelsList extends makeSchema("github.com/sashabaranov/go-openai/ModelsList", {
    data: ArrayOf(Model),
}) {}
