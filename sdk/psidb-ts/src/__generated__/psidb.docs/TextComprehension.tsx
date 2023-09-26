import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";


export class TextComprehension extends makeSchema("psidb.docs/TextComprehension", {
    content: PrimitiveTypes.String,
    currentIndex: PrimitiveTypes.Float64,
    linesPerChunk: PrimitiveTypes.Float64,
    linesPerPage: PrimitiveTypes.Float64,
    name: PrimitiveTypes.String,
    observations: ArrayOf(PrimitiveTypes.String),
}) {}
