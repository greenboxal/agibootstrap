import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { error } from "@psidb/psidb-sdk/types//error";
import { bool } from "@psidb/psidb-sdk/types//bool";


export const ReadableStorage = makeInterface({
    name: "github.com/ipld/go-ipld-prime/storage/ReadableStorage",
    methods: {
        Get: PrimitiveTypes.Func(Context, PrimitiveTypes.String)(PrimitiveTypes.Array(uint8)(error)),
        Has: PrimitiveTypes.Func(Context, PrimitiveTypes.String)(bool, error),
    },
});
