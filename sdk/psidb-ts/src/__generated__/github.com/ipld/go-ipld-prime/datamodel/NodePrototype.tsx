import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { NodeBuilder } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeBuilder";


export const NodePrototype = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/NodePrototype",
    methods: {
        NewBuilder: PrimitiveTypes.Func()(NodeBuilder),
    },
});
