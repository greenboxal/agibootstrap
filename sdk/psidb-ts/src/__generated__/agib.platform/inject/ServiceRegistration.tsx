import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { ServiceDefinition } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceDefinition";
import { ResolutionContext } from "@psidb/psidb-sdk/types/agib.platform/inject/ResolutionContext";
import { error } from "@psidb/psidb-sdk/types//error";
import { ServiceKey } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceKey";


export const ServiceRegistration = makeInterface({
    name: "agib.platform/inject/ServiceRegistration",
    methods: {
        GetDefinition: PrimitiveTypes.Func()(ServiceDefinition),
        GetInstance: PrimitiveTypes.Func(ResolutionContext)(PrimitiveTypes.Any, error),
        GetKey: PrimitiveTypes.Func()(ServiceKey),
    },
});
