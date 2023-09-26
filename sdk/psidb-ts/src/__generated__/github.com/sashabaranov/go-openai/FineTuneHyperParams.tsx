import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FineTuneHyperParams extends makeSchema("github.com/sashabaranov/go-openai/FineTuneHyperParams", {
    batch_size: PrimitiveTypes.Float64,
    learning_rate_multiplier: PrimitiveTypes.Float64,
    n_epochs: PrimitiveTypes.Float64,
    prompt_loss_weight: PrimitiveTypes.Float64,
}) {}
