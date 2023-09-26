import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { BasicListChangeEvent } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/BasicListChangeEvent";


export const BasicListListener = makeInterface({
    name: "agib.platform/stdlib/obsfx/collectionsfx/BasicListListener",
    methods: {
        OnListChangedRaw: PrimitiveTypes.Func(BasicListChangeEvent)(),
    },
});
