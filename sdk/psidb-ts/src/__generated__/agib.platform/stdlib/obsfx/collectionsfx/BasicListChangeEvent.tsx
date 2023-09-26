import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { BasicObservableList } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/BasicObservableList";
import { bool } from "@psidb/psidb-sdk/types//bool";


export const BasicListChangeEvent = makeInterface({
    name: "agib.platform/stdlib/obsfx/collectionsfx/BasicListChangeEvent",
    methods: {
        AddedCount: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        BasicList: PrimitiveTypes.Func()(BasicObservableList),
        From: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        GetPermutation: PrimitiveTypes.Func(PrimitiveTypes.Integer)(PrimitiveTypes.Integer),
        Next: PrimitiveTypes.Func()(bool),
        Permutations: PrimitiveTypes.Func()(PrimitiveTypes.Array(PrimitiveTypes.Integer)),
        RemovedCount: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Reset: PrimitiveTypes.Func()(),
        To: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        WasAdded: PrimitiveTypes.Func()(bool),
        WasPermutated: PrimitiveTypes.Func()(bool),
        WasRemoved: PrimitiveTypes.Func()(bool),
        WasReplaced: PrimitiveTypes.Func()(bool),
        WasUpdated: PrimitiveTypes.Func()(bool),
    },
});
