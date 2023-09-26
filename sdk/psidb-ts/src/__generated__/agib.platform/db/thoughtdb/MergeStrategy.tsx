import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Repo } from "@psidb/psidb-sdk/types/agib.platform/db/thoughtdb/Repo";
import { Branch } from "@psidb/psidb-sdk/types/agib.platform/db/thoughtdb/Branch";
import { error } from "@psidb/psidb-sdk/types//error";


export const MergeStrategy = makeInterface({
    name: "agib.platform/db/thoughtdb/MergeStrategy",
    methods: {
        Merge: PrimitiveTypes.Func(Context, PrimitiveTypes.Pointer(Repo)(Branch, PrimitiveTypes.Array(Branch)))(error),
    },
});
