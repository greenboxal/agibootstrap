import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { error } from "@psidb/psidb-sdk/types//error";


export const MapIterator = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/MapIterator",
    methods: {
        Done: PrimitiveTypes.Func()(bool),
        Next: PrimitiveTypes.Func()(Node, Node, error),
    },
});
