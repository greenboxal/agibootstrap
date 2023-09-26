import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class RootNode extends makeSchema("stdlib/RootNode", {
    UUID: PrimitiveTypes.String,
}) {}
