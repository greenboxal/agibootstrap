import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class General extends makeSchema("gonum.org/v1/gonum/blas/blas64/General", {
    Cols: PrimitiveTypes.Float64,
    Data: ArrayOf(PrimitiveTypes.Float64),
    Rows: PrimitiveTypes.Float64,
    Stride: PrimitiveTypes.Float64,
}) {}
