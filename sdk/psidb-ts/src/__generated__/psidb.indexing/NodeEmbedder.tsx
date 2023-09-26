import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { GraphEmbedding } from "@psidb/psidb-sdk/types/psidb.indexing/GraphEmbedding";
import { error } from "@psidb/psidb-sdk/types//error";


export const NodeEmbedder = makeInterface({
    name: "psidb.indexing/NodeEmbedder",
    methods: {
        Dimensions: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        EmbeddingsForNode: PrimitiveTypes.Func(Context, Node)(Iterator(GraphEmbedding)(error)),
    },
});
