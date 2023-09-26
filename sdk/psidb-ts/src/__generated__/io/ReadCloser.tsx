import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export const ReadCloser = makeInterface({
    name: "io/ReadCloser",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Read: PrimitiveTypes.Func(PrimitiveTypes.Array(uint8))(PrimitiveTypes.Integer, error),
    },
});
