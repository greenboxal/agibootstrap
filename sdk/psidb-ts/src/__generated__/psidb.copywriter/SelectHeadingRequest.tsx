import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class SelectHeadingRequest extends makeSchema("psidb.copywriter/SelectHeadingRequest", {
    heading: PrimitiveTypes.String,
    heading_level: PrimitiveTypes.Float64,
}) {}
