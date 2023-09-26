import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Agent extends makeSchema("psidb.agents/Agent", {
    last_message: PrimitiveTypes.String,
    name: PrimitiveTypes.String,
}) {}
