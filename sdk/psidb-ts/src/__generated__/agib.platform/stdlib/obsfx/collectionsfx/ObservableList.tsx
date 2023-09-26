import { Type } from "@psidb/psidb-sdk/types/reflect/Type";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { BasicListListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/BasicListListener";
import { ListListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/ListListener";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { InvalidationListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/InvalidationListener";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Iterator } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/Iterator";


export function ObservableList<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/obsfx/collectionsfx/ObservableList(github.com/greenboxal/agibootstrap/psidb/psi/Node)",
        methods: {
            AddBasicListListener: PrimitiveTypes.Func(BasicListListener)(),
            AddListListener: PrimitiveTypes.Func(ListListener(Node))(),
            AddListener: PrimitiveTypes.Func(InvalidationListener)(),
            Contains: PrimitiveTypes.Func(Node)(bool),
            Get: PrimitiveTypes.Func(PrimitiveTypes.Integer)(Node),
            IndexOf: PrimitiveTypes.Func(Node)(PrimitiveTypes.Integer),
            Iterator: PrimitiveTypes.Func()(Iterator(Node)),
            Len: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
            RawGet: PrimitiveTypes.Func(PrimitiveTypes.Integer)(PrimitiveTypes.Any),
            RemoveBasicListListener: PrimitiveTypes.Func(BasicListListener)(),
            RemoveListListener: PrimitiveTypes.Func(ListListener(Node))(),
            RemoveListener: PrimitiveTypes.Func(InvalidationListener)(),
            RuntimeElementType: PrimitiveTypes.Func()(Type),
            Slice: PrimitiveTypes.Func()(PrimitiveTypes.Array(Node)),
            SubSlice: PrimitiveTypes.Func(PrimitiveTypes.Integer, PrimitiveTypes.Integer)(PrimitiveTypes.Array(Node)),
        },
    })
}