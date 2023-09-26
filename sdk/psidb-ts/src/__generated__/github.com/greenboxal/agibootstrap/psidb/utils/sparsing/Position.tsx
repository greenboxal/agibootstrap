import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Position extends makeSchema("github.com/greenboxal/agibootstrap/psidb/utils/sparsing/Position", {
    Column: PrimitiveTypes.Float64,
    Line: PrimitiveTypes.Float64,
    Offset: PrimitiveTypes.Float64,
}) {}
