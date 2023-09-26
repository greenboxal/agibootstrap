import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Vector extends makeSchema("gonum.org/v1/gonum/blas/blas64/Vector", {
    Data: ArrayOf(PrimitiveTypes.Float64),
    Inc: PrimitiveTypes.Float64,
    N: PrimitiveTypes.Float64,
}) {}
