import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uintptr } from "@psidb/psidb-sdk/types//uintptr";
import { error } from "@psidb/psidb-sdk/types//error";
import { bool } from "@psidb/psidb-sdk/types//bool";


export const RawConn = makeInterface({
    name: "syscall/RawConn",
    methods: {
        Control: PrimitiveTypes.Func(PrimitiveTypes.Func(uintptr)())(error),
        Read: PrimitiveTypes.Func(PrimitiveTypes.Func(uintptr)(bool))(error),
        Write: PrimitiveTypes.Func(PrimitiveTypes.Func(uintptr)(bool))(error),
    },
});
