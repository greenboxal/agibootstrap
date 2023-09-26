import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Matrix } from "@psidb/psidb-sdk/types/gonum.org/v1/gonum/mat/Matrix";


export const Vector = makeInterface({
    name: "gonum.org/v1/gonum/mat/Vector",
    methods: {
        At: PrimitiveTypes.Func(PrimitiveTypes.Integer, PrimitiveTypes.Integer)(PrimitiveTypes.Float64),
        AtVec: PrimitiveTypes.Func(PrimitiveTypes.Integer)(PrimitiveTypes.Float64),
        Dims: PrimitiveTypes.Func()(PrimitiveTypes.Integer, PrimitiveTypes.Integer),
        Len: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        T: PrimitiveTypes.Func()(Matrix),
    },
});
