import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class TextComprehensionStepRequest extends makeSchema("psidb.docs/TextComprehensionStepRequest", {
    currentSteps: PrimitiveTypes.Float64,
    maxSteps: PrimitiveTypes.Float64,
}) {}
