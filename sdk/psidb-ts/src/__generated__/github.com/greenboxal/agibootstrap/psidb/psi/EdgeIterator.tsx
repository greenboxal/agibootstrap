import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Edge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Edge";


export const EdgeIterator = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/EdgeIterator",
    methods: {
        Next: PrimitiveTypes.Func()(bool),
        Value: PrimitiveTypes.Func()(Edge),
    },
});
