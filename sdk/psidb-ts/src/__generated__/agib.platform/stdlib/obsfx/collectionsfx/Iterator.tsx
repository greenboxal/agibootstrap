import { Type, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { bool } from "@psidb/psidb-sdk/types//bool";


export function Iterator<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/obsfx/collectionsfx/Iterator(github.com/greenboxal/agibootstrap/psidb/psi/Node)",
        methods: {
            Item: PrimitiveTypes.Func()(Node),
            Next: PrimitiveTypes.Func()(bool),
            Reset: PrimitiveTypes.Func()(),
            Value: PrimitiveTypes.Func()(Node),
        },
    })
}