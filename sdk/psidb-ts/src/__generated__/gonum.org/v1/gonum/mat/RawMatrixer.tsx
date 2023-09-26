import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { General } from "@psidb/psidb-sdk/types/gonum.org/v1/gonum/blas/blas64/General";


export const RawMatrixer = makeInterface({
    name: "gonum.org/v1/gonum/mat/RawMatrixer",
    methods: {
        RawMatrix: PrimitiveTypes.Func()(General),
    },
});
