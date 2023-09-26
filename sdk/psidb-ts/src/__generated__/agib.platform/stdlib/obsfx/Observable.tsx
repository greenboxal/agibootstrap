import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { InvalidationListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/InvalidationListener";


export const Observable = makeInterface({
    name: "agib.platform/stdlib/obsfx/Observable",
    methods: {
        AddListener: PrimitiveTypes.Func(InvalidationListener)(),
        RemoveListener: PrimitiveTypes.Func(InvalidationListener)(),
    },
});
