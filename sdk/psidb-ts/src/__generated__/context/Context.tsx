import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Time } from "@psidb/psidb-sdk/types/time/Time";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { chan } from "@psidb/psidb-sdk/types//chan";
import { error } from "@psidb/psidb-sdk/types//error";


export const Context = makeInterface({
    name: "context/Context",
    methods: {
        Deadline: PrimitiveTypes.Func()(Time, bool),
        Done: PrimitiveTypes.Func()(chan),
        Err: PrimitiveTypes.Func()(error),
        Value: PrimitiveTypes.Func(PrimitiveTypes.Any)(PrimitiveTypes.Any),
    },
});
