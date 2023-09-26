import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class Symbol extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/analysis/Symbol", {
    is_local: PrimitiveTypes.Boolean,
    is_resolved: PrimitiveTypes.Boolean,
    name: PrimitiveTypes.String,
    reference_distance: PrimitiveTypes.Float64,
    root_distance: PrimitiveTypes.Float64,
    scope_distance: PrimitiveTypes.Float64,
}) {}
