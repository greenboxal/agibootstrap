import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { NodeEmbedder } from "@psidb/psidb-sdk/types/psidb.indexing/NodeEmbedder";
import { BasicIndex } from "@psidb/psidb-sdk/types/psidb.indexing/BasicIndex";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { SearchRequest } from "@psidb/psidb-sdk/types/psidb.indexing/SearchRequest";
import { NodeSearchHit } from "@psidb/psidb-sdk/types/psidb.indexing/NodeSearchHit";


export const NodeIndex = makeInterface({
    name: "psidb.indexing/NodeIndex",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        Embedder: PrimitiveTypes.Func()(NodeEmbedder),
        Index: PrimitiveTypes.Func()(BasicIndex),
        IndexNode: PrimitiveTypes.Func(Context, Node)(error),
        Search: PrimitiveTypes.Func(Context, SearchRequest)(Iterator(NodeSearchHit)(error)),
    },
});
