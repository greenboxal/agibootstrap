import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";


export const TaskProgress = makeInterface({
    name: "agib.platform/tasks/TaskProgress",
    methods: {
        Context: PrimitiveTypes.Func()(Context),
        Update: PrimitiveTypes.Func(PrimitiveTypes.Integer, PrimitiveTypes.Integer)(),
    },
});
