import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Type } from "@psidb/psidb-sdk/types/reflect/Type";


export class StructField extends makeSchema("reflect/StructField", {
    Anonymous: PrimitiveTypes.Boolean,
    Index: ArrayOf(PrimitiveTypes.Integer),
    Name: PrimitiveTypes.String,
    Offset: PrimitiveTypes.Float64,
    PkgPath: PrimitiveTypes.String,
    Tag: PrimitiveTypes.String,
    Type: Type,
}) {}
