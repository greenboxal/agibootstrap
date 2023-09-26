import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { FileInfo } from "@psidb/psidb-sdk/types/io/fs/FileInfo";
import { error } from "@psidb/psidb-sdk/types//error";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { FileMode } from "@psidb/psidb-sdk/types/io/fs/FileMode";


export const DirEntry = makeInterface({
    name: "io/fs/DirEntry",
    methods: {
        Info: PrimitiveTypes.Func()(FileInfo, error),
        IsDir: PrimitiveTypes.Func()(bool),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        Type: PrimitiveTypes.Func()(FileMode),
    },
});
