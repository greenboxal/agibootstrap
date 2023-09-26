import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { CommandSheetLine } from "@psidb/psidb-sdk/types/psidb.jukebox/CommandSheetLine";
import { IToken } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/utils/sparsing/IToken";


export class CommandSheetNode extends makeSchema("psidb.jukebox/CommandSheetNode", {
    Lines: ArrayOf(CommandSheetLine),
    end: IToken,
    start: IToken,
}) {}
