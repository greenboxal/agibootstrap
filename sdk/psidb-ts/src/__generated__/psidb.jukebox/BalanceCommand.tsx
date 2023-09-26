import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class BalanceCommand extends makeSchema("psidb.jukebox/BalanceCommand", {
    balance: PrimitiveTypes.Float64,
}) {}
