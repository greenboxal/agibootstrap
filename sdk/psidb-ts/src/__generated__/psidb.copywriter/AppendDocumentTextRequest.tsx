import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class AppendDocumentTextRequest extends makeSchema("psidb.copywriter/AppendDocumentTextRequest", {
    heading: PrimitiveTypes.String,
    heading_level: PrimitiveTypes.Float64,
    markdown: PrimitiveTypes.String,
}) {}
