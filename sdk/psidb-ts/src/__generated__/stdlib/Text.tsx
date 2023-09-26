import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Text extends makeSchema("stdlib/Text", {
    value: PrimitiveTypes.String,
}) {}
