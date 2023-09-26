import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Package extends makeSchema("psidb.typing/Package", {
    name: PrimitiveTypes.String,
}) {}
