import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Command } from "@psidb/psidb-sdk/types/psidb.jukebox/Command";


export class CommandSheet extends makeSchema("psidb.jukebox/CommandSheet", {
    Commands: ArrayOf(Command),
}) {}
