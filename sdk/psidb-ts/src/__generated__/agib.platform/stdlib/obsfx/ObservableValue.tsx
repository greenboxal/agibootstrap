import { Type, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { ChangeListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/ChangeListener";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { InvalidationListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/InvalidationListener";


export function ObservableValue<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/obsfx/ObservableValue(github.com/greenboxal/agibootstrap/psidb/psi/Node)",
        methods: {
            AddChangeListener: PrimitiveTypes.Func(ChangeListener(Node))(),
            AddListener: PrimitiveTypes.Func(InvalidationListener)(),
            RawValue: PrimitiveTypes.Func()(PrimitiveTypes.Any),
            RemoveChangeListener: PrimitiveTypes.Func(ChangeListener(Node))(),
            RemoveListener: PrimitiveTypes.Func(InvalidationListener)(),
            Value: PrimitiveTypes.Func()(Node),
        },
    })
}