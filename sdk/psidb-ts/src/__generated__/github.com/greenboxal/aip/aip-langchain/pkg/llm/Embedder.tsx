import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Embedding } from "@psidb/psidb-sdk/types/github.com/greenboxal/aip/aip-langchain/pkg/llm/Embedding";
import { error } from "@psidb/psidb-sdk/types//error";


export const Embedder = makeInterface({
    name: "github.com/greenboxal/aip/aip-langchain/pkg/llm/Embedder",
    methods: {
        Dimensions: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        GetEmbeddings: PrimitiveTypes.Func(Context, PrimitiveTypes.Array(PrimitiveTypes.String))(PrimitiveTypes.Array(Embedding)(error)),
        Identity: PrimitiveTypes.Func()(PrimitiveTypes.String),
        MaxTokensPerChunk: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
    },
});
