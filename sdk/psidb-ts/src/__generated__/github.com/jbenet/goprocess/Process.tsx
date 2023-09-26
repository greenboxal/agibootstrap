import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { chan } from "@psidb/psidb-sdk/types//chan";
import { ProcessFunc } from "@psidb/psidb-sdk/types/github.com/jbenet/goprocess/ProcessFunc";
import { TeardownFunc } from "@psidb/psidb-sdk/types/github.com/jbenet/goprocess/TeardownFunc";

const _F = {} as any

export const Process = makeInterface({
    name: "github.com/jbenet/goprocess/Process",
    methods: {
        AddChild: PrimitiveTypes.Func(_F["Process"])(),
        AddChildNoWait: PrimitiveTypes.Func(_F["Process"])(),
        Close: PrimitiveTypes.Func()(error),
        CloseAfterChildren: PrimitiveTypes.Func()(error),
        Closed: PrimitiveTypes.Func()(chan),
        Closing: PrimitiveTypes.Func()(chan),
        Err: PrimitiveTypes.Func()(error),
        Go: PrimitiveTypes.Func(ProcessFunc(_F["Process"]))(_F["Process"]),
        SetTeardown: PrimitiveTypes.Func(TeardownFunc(error))(),
        WaitFor: PrimitiveTypes.Func(_F["Process"])(),
    },
});
_F["Process"] = Process;
