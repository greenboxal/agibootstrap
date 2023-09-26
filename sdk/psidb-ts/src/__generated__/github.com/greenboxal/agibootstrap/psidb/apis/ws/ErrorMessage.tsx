import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export class ErrorMessage extends makeSchema("github.com/greenboxal/agibootstrap/psidb/apis/ws/ErrorMessage", {
    error: PrimitiveTypes.String,
}) {}
