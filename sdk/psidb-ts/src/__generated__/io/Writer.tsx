import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { error } from "@psidb/psidb-sdk/types//error";


export const Writer = makeInterface({
    name: "io/Writer",
    methods: {
        Write: PrimitiveTypes.Func(PrimitiveTypes.Array(uint8))(PrimitiveTypes.Integer, error),
    },
});
