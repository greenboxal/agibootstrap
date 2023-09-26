import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class PathElement extends makeSchema("github.com/greenboxal/agibootstrap/psidb/psi/PathElement", {
    Index: PrimitiveTypes.Float64,
    Kind: PrimitiveTypes.String,
    Name: PrimitiveTypes.String,
}) {}
