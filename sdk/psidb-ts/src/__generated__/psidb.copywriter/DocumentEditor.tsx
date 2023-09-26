import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { DocumentHeading } from "@psidb/psidb-sdk/types/psidb.copywriter/DocumentHeading";


export class DocumentEditor extends makeSchema("psidb.copywriter/DocumentEditor", {
    last_section: PrimitiveTypes.Float64,
    name: PrimitiveTypes.String,
    sections: ArrayOf(DocumentHeading),
}) {}
