import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { BasicListListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/BasicListListener";
import { InvalidationListener } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/InvalidationListener";
import { Type } from "@psidb/psidb-sdk/types/reflect/Type";


export const BasicObservableList = makeInterface({
    name: "agib.platform/stdlib/obsfx/collectionsfx/BasicObservableList",
    methods: {
        AddBasicListListener: PrimitiveTypes.Func(BasicListListener)(),
        AddListener: PrimitiveTypes.Func(InvalidationListener)(),
        Len: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        RawGet: PrimitiveTypes.Func(PrimitiveTypes.Integer)(PrimitiveTypes.Any),
        RemoveBasicListListener: PrimitiveTypes.Func(BasicListListener)(),
        RemoveListener: PrimitiveTypes.Func(InvalidationListener)(),
        RuntimeElementType: PrimitiveTypes.Func()(Type),
    },
});
