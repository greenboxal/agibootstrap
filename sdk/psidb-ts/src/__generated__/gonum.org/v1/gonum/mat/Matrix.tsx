import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";

const _F = {} as any

export const Matrix = makeInterface({
    name: "gonum.org/v1/gonum/mat/Matrix",
    methods: {
        At: PrimitiveTypes.Func(PrimitiveTypes.Integer, PrimitiveTypes.Integer)(PrimitiveTypes.Float64),
        Dims: PrimitiveTypes.Func()(PrimitiveTypes.Integer, PrimitiveTypes.Integer),
        T: PrimitiveTypes.Func()(_F["Matrix"]),
    },
});
_F["Matrix"] = Matrix;
