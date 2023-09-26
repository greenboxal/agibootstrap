import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class FineTuneRequest extends makeSchema("github.com/sashabaranov/go-openai/FineTuneRequest", {
    batch_size: PrimitiveTypes.Float64,
    classification_betas: ArrayOf(PrimitiveTypes.Float32),
    classification_n_classes: PrimitiveTypes.Float64,
    classification_positive_class: PrimitiveTypes.String,
    compute_classification_metrics: PrimitiveTypes.Boolean,
    learning_rate_multiplier: PrimitiveTypes.Float64,
    model: PrimitiveTypes.String,
    n_epochs: PrimitiveTypes.Float64,
    prompt_loss_rate: PrimitiveTypes.Float64,
    suffix: PrimitiveTypes.String,
    training_file: PrimitiveTypes.String,
    validation_file: PrimitiveTypes.String,
}) {}
