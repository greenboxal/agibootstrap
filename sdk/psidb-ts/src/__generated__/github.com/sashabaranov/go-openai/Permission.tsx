import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Permission extends makeSchema("github.com/sashabaranov/go-openai/Permission", {
    allow_create_engine: PrimitiveTypes.Boolean,
    allow_fine_tuning: PrimitiveTypes.Boolean,
    allow_logprobs: PrimitiveTypes.Boolean,
    allow_sampling: PrimitiveTypes.Boolean,
    allow_search_indices: PrimitiveTypes.Boolean,
    allow_view: PrimitiveTypes.Boolean,
    created: PrimitiveTypes.Float64,
    group: PrimitiveTypes.Any,
    id: PrimitiveTypes.String,
    is_blocking: PrimitiveTypes.Boolean,
    object: PrimitiveTypes.String,
    organization: PrimitiveTypes.String,
}) {}
