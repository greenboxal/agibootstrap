import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { PromptResponseChoice } from "@psidb/psidb-sdk/types/psidb.agents/PromptResponseChoice";
import { error } from "@psidb/psidb-sdk/types//error";


export const ResultParser = makeInterface({
    name: "psidb.agents/ResultParser",
    methods: {
        ParseChoice: PrimitiveTypes.Func(Context, PromptResponseChoice)(error),
    },
});
