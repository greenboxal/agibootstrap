import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { ServiceKey } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceKey";
import { Type } from "@psidb/psidb-sdk/types/reflect/Type";


export class ServiceDefinition extends makeSchema("agib.platform/inject/ServiceDefinition", {
    Dependencies: ArrayOf(ServiceKey),
    Key: makeSchema("", {
        Name: PrimitiveTypes.String,
        Type: Type,
    }),
}) {}
