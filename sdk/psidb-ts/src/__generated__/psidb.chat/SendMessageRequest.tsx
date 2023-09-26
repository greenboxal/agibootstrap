import { makeSchema, ArrayOf, PrimitiveTypes, MapOf } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class SendMessageRequest extends makeSchema("psidb.chat/SendMessageRequest", {
    attachments: ArrayOf(Path),
    from: makeSchema("", {
        id: PrimitiveTypes.String,
        name: PrimitiveTypes.String,
        role: PrimitiveTypes.String,
    }),
    function: PrimitiveTypes.String,
    function_arguments: PrimitiveTypes.String,
    metadata: MapOf(PrimitiveTypes.String, PrimitiveTypes.Any),
    model_options: makeSchema("", {
        force_function_call: PrimitiveTypes.String,
        frequency_penalty: PrimitiveTypes.Float32,
        logit_bias: MapOf(PrimitiveTypes.String, PrimitiveTypes.Integer),
        max_tokens: PrimitiveTypes.Integer,
        model: PrimitiveTypes.String,
        presence_penalty: PrimitiveTypes.Float32,
        stop: ArrayOf(PrimitiveTypes.String),
        temperature: PrimitiveTypes.Float32,
        top_p: PrimitiveTypes.Float32,
    }),
    text: PrimitiveTypes.String,
    timestamp: PrimitiveTypes.String,
}) {}
