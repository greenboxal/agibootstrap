import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { FineTuneEvent } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/FineTuneEvent";
import { File } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/File";


export class FineTune extends makeSchema("github.com/sashabaranov/go-openai/FineTune", {
    created_at: PrimitiveTypes.Float64,
    events: ArrayOf(FineTuneEvent),
    fine_tuned_model: PrimitiveTypes.String,
    hyperparams: makeSchema("", {
        batch_size: PrimitiveTypes.Float64,
        learning_rate_multiplier: PrimitiveTypes.Float64,
        n_epochs: PrimitiveTypes.Float64,
        prompt_loss_weight: PrimitiveTypes.Float64,
    }),
    id: PrimitiveTypes.String,
    model: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
    organization_id: PrimitiveTypes.String,
    result_files: ArrayOf(File),
    status: PrimitiveTypes.String,
    training_files: ArrayOf(File),
    updated_at: PrimitiveTypes.Float64,
    validation_files: ArrayOf(File),
}) {}
