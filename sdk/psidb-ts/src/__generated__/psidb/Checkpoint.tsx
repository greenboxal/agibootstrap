import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { bool } from "@psidb/psidb-sdk/types//bool";


export const Checkpoint = makeInterface({
    name: "psidb/Checkpoint",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Get: PrimitiveTypes.Func()(PrimitiveTypes.UnsignedInteger, bool, error),
        Update: PrimitiveTypes.Func(PrimitiveTypes.UnsignedInteger, bool)(error),
    },
});
