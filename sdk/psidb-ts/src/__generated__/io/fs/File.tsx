import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { FileInfo } from "@psidb/psidb-sdk/types/io/fs/FileInfo";


export const File = makeInterface({
    name: "io/fs/File",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Read: PrimitiveTypes.Func(PrimitiveTypes.Array(uint8))(PrimitiveTypes.Integer, error),
        Stat: PrimitiveTypes.Func()(FileInfo, error),
    },
});
