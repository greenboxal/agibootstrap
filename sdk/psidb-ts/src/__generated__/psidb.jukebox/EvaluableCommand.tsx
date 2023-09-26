import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { CommandContext } from "@psidb/psidb-sdk/types/psidb.jukebox/CommandContext";
import { error } from "@psidb/psidb-sdk/types//error";


export const EvaluableCommand = makeInterface({
    name: "psidb.jukebox/EvaluableCommand",
    methods: {
        Evaluate: PrimitiveTypes.Func(PrimitiveTypes.Pointer(CommandContext))(error),
    },
});
