import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Permission } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/Permission";


export class Model extends makeSchema("github.com/sashabaranov/go-openai/Model", {
    created: PrimitiveTypes.Float64,
    id: PrimitiveTypes.String,
    object: PrimitiveTypes.String,
    owned_by: PrimitiveTypes.String,
    parent: PrimitiveTypes.String,
    permission: ArrayOf(Permission),
    root: PrimitiveTypes.String,
}) {}
