import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Command } from "@psidb/psidb-sdk/types/psidb.jukebox/Command";
import { IToken } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/utils/sparsing/IToken";


export class CommandSheetLine extends makeSchema("psidb.jukebox/CommandSheetLine", {
    Commands: ArrayOf(Command),
    end: IToken,
    start: IToken,
}) {}
