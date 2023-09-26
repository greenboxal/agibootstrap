import { Type, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { BasicObservableList } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/BasicObservableList";
import { ObservableList } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/ObservableList";
import { bool } from "@psidb/psidb-sdk/types//bool";


export function ListChangeEvent<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/obsfx/collectionsfx/ListChangeEvent(github.com/greenboxal/agibootstrap/psidb/psi/Node)",
        methods: {
            AddedCount: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
            AddedSlice: PrimitiveTypes.Func()(PrimitiveTypes.Array(Node)),
            BasicList: PrimitiveTypes.Func()(BasicObservableList),
            From: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
            GetPermutation: PrimitiveTypes.Func(PrimitiveTypes.Integer)(PrimitiveTypes.Integer),
            List: PrimitiveTypes.Func()(ObservableList(Node)),
            Next: PrimitiveTypes.Func()(bool),
            Permutations: PrimitiveTypes.Func()(PrimitiveTypes.Array(PrimitiveTypes.Integer)),
            RemovedCount: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
            RemovedSlice: PrimitiveTypes.Func()(PrimitiveTypes.Array(Node)),
            Reset: PrimitiveTypes.Func()(),
            To: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
            WasAdded: PrimitiveTypes.Func()(bool),
            WasPermutated: PrimitiveTypes.Func()(bool),
            WasRemoved: PrimitiveTypes.Func()(bool),
            WasReplaced: PrimitiveTypes.Func()(bool),
            WasUpdated: PrimitiveTypes.Func()(bool),
        },
    })
}