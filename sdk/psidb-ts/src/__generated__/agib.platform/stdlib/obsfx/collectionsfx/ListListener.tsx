import { Type, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { ListChangeEvent } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/ListChangeEvent";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export function ListListener<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/obsfx/collectionsfx/ListListener(github.com/greenboxal/agibootstrap/psidb/psi/Node)",
        methods: {
            OnListChanged: PrimitiveTypes.Func(ListChangeEvent(Node))(),
        },
    })
}