import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class FunctionCall extends makeSchema("psidb.chat/FunctionCall", {
    arguments: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
}) {}
