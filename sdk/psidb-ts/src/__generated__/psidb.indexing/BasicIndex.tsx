import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { IndexNodeRequest } from "@psidb/psidb-sdk/types/psidb.indexing/IndexNodeRequest";
import { IndexedItem } from "@psidb/psidb-sdk/types/psidb.indexing/IndexedItem";
import { SearchRequest } from "@psidb/psidb-sdk/types/psidb.indexing/SearchRequest";
import { BasicSearchHit } from "@psidb/psidb-sdk/types/psidb.indexing/BasicSearchHit";


export const BasicIndex = makeInterface({
    name: "psidb.indexing/BasicIndex",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Dimensions: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        IndexNode: PrimitiveTypes.Func(Context, IndexNodeRequest)(IndexedItem, error),
        Load: PrimitiveTypes.Func()(error),
        Rebuild: PrimitiveTypes.Func(Context)(error),
        Save: PrimitiveTypes.Func()(error),
        Search: PrimitiveTypes.Func(Context, SearchRequest)(Iterator(BasicSearchHit)(error)),
        Truncate: PrimitiveTypes.Func(Context)(error),
    },
});
