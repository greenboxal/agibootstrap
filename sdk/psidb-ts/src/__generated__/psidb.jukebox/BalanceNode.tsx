import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class BalanceNode extends makeSchema("psidb.jukebox/BalanceNode", {
    Balance: PrimitiveTypes.Float64,
}) {}
