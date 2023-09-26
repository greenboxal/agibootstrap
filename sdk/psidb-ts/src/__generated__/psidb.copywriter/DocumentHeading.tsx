import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class DocumentHeading extends makeSchema("psidb.copywriter/DocumentHeading", {
    content: PrimitiveTypes.String,
    heading: PrimitiveTypes.String,
    heading_level: PrimitiveTypes.Float64,
    index: PrimitiveTypes.Float64,
    order: PrimitiveTypes.Float64,
    parent_section: PrimitiveTypes.Float64,
}) {}
