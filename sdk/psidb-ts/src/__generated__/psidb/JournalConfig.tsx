import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Journal } from "@psidb/psidb-sdk/types/psidb/Journal";
import { error } from "@psidb/psidb-sdk/types//error";


export const JournalConfig = makeInterface({
    name: "psidb/JournalConfig",
    methods: {
        CreateJournal: PrimitiveTypes.Func(Context)(Journal, error),
    },
});
