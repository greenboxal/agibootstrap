import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { error } from "@psidb/psidb-sdk/types//error";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export const WritableStorage = makeInterface({
    name: "github.com/ipld/go-ipld-prime/storage/WritableStorage",
    methods: {
        Has: PrimitiveTypes.Func(Context, PrimitiveTypes.String)(bool, error),
        Put: PrimitiveTypes.Func(Context, PrimitiveTypes.String, PrimitiveTypes.Array(uint8))(error),
    },
});
