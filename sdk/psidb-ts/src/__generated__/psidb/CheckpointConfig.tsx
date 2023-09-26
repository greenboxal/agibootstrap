import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Checkpoint } from "@psidb/psidb-sdk/types/psidb/Checkpoint";
import { error } from "@psidb/psidb-sdk/types//error";


export const CheckpointConfig = makeInterface({
    name: "psidb/CheckpointConfig",
    methods: {
        CreateCheckpoint: PrimitiveTypes.Func(Context)(Checkpoint, error),
    },
});
