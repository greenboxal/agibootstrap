import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FineTune } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FineTune";


export class FineTuneList extends makeSchema("github.com/sashabaranov/go-openai/FineTuneList", {
    data: ArrayOf(FineTune),
    object: PrimitiveTypes.String,
}) {}
