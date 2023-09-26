import { makeSchema, MapOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PromptBuilder extends makeSchema("psidb.agents/PromptBuilder", {
    Context: MapOf(PrimitiveTypes.String, PrimitiveTypes.Any),
}) {}
