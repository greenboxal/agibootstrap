import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FineTuneEvent } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FineTuneEvent";


export class FineTuneEventList extends makeSchema("github.com/sashabaranov/go-openai/FineTuneEventList", {
    data: ArrayOf(FineTuneEvent),
    object: PrimitiveTypes.String,
}) {}
