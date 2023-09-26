import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { MountDefinition } from "@psidb/psidb-sdk/types/psidb/MountDefinition";
import { error } from "@psidb/psidb-sdk/types//error";


export const MountTarget = makeInterface({
    name: "psidb/MountTarget",
    methods: {
        Mount: PrimitiveTypes.Func(Context, MountDefinition)(PrimitiveTypes.Any, error),
    },
});
