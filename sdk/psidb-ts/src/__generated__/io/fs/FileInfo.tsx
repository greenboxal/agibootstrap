import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Time } from "@psidb/psidb-sdk/types/time/Time";
import { FileMode } from "@psidb/psidb-sdk/types/io/fs/FileMode";


export const FileInfo = makeInterface({
    name: "io/fs/FileInfo",
    methods: {
        IsDir: PrimitiveTypes.Func()(bool),
        ModTime: PrimitiveTypes.Func()(Time),
        Mode: PrimitiveTypes.Func()(FileMode),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        Size: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Sys: PrimitiveTypes.Func()(PrimitiveTypes.Any),
    },
});
