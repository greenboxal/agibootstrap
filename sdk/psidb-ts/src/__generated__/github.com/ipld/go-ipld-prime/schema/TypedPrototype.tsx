import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { NodeBuilder } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeBuilder";
import { NodePrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodePrototype";
import { Type } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/schema/Type";


export const TypedPrototype = makeInterface({
    name: "github.com/ipld/go-ipld-prime/schema/TypedPrototype",
    methods: {
        NewBuilder: PrimitiveTypes.Func()(NodeBuilder),
        Representation: PrimitiveTypes.Func()(NodePrototype),
        Type: PrimitiveTypes.Func()(Type),
    },
});
