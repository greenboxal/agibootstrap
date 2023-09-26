import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Type } from "@psidb/psidb-sdk/types/reflect/Type";


export class Method extends makeSchema("reflect/Method", {
    Func: makeSchema("", {
    }),
    Index: PrimitiveTypes.Float64,
    Name: PrimitiveTypes.String,
    PkgPath: PrimitiveTypes.String,
    Type: Type,
}) {}
