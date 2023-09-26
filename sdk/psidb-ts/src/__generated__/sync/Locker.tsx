import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export const Locker = makeInterface({
    name: "sync/Locker",
    methods: {
        Lock: PrimitiveTypes.Func()(),
        Unlock: PrimitiveTypes.Func()(),
    },
});
