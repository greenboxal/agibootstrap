import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Type } from "@psidb/psidb-sdk/types/reflect/Type";


export class ServiceKey extends makeSchema("agib.platform/inject/ServiceKey", {
    Name: PrimitiveTypes.String,
    Type: Type,
}) {}
