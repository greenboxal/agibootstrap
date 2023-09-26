import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { DataStore } from "@psidb/psidb-sdk/types/psidb/DataStore";
import { error } from "@psidb/psidb-sdk/types//error";


export const DataStoreConfig = makeInterface({
    name: "psidb/DataStoreConfig",
    methods: {
        CreateDataStore: PrimitiveTypes.Func(Context)(DataStore, error),
    },
});
