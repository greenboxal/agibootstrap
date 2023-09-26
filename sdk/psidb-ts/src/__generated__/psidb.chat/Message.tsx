import { makeSchema, ArrayOf, PrimitiveTypes, MapOf } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";
import { FunctionCall } from "@psidb/psidb-sdk/types/psidb.chat/FunctionCall";


export class Message extends makeSchema("psidb.chat/Message", {
    attachments: ArrayOf(Path),
    from: makeSchema("", {
        id: PrimitiveTypes.String,
        name: PrimitiveTypes.String,
        role: PrimitiveTypes.String,
    }),
    function_call: FunctionCall,
    kind: PrimitiveTypes.String,
    metadata: MapOf(PrimitiveTypes.String, PrimitiveTypes.Any),
    reply_to: ArrayOf(Path),
    text: PrimitiveTypes.String,
    timestamp: PrimitiveTypes.String,
}) {}

