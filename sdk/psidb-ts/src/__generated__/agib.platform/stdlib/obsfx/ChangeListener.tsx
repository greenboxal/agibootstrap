import { Type, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export function ChangeListener<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/obsfx/ChangeListener(github.com/greenboxal/agibootstrap/psidb/psi/Node)",
        methods: {
            OnChanged: PrimitiveTypes.Func(ObservableValue(Node)(Node, Node))(),
        },
    })
}