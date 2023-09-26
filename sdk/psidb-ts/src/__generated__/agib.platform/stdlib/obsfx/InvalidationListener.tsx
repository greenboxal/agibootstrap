import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Observable } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/Observable";


export const InvalidationListener = makeInterface({
    name: "agib.platform/stdlib/obsfx/InvalidationListener",
    methods: {
        OnInvalidated: PrimitiveTypes.Func(Observable)(),
    },
});
