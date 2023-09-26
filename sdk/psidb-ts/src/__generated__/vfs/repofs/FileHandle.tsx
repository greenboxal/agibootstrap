import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { ReadCloser } from "@psidb/psidb-sdk/types/io/ReadCloser";
import { Reader } from "@psidb/psidb-sdk/types/io/Reader";


export const FileHandle = makeInterface({
    name: "vfs/repofs/FileHandle",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Get: PrimitiveTypes.Func()(ReadCloser, error),
        Put: PrimitiveTypes.Func(Reader)(error),
    },
});
