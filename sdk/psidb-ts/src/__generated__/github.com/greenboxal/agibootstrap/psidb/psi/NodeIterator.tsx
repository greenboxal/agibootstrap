import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export const NodeIterator = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/NodeIterator",
    methods: {
        Next: PrimitiveTypes.Func()(bool),
        Value: PrimitiveTypes.Func()(Node),
    },
});
