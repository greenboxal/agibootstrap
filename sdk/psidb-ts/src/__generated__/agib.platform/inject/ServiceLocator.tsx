import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { ServiceKey } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceKey";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { ServiceRegistration } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceRegistration";
import { error } from "@psidb/psidb-sdk/types//error";


export const ServiceLocator = makeInterface({
    name: "agib.platform/inject/ServiceLocator",
    methods: {
        GetRegistration: PrimitiveTypes.Func(ServiceKey, bool)(ServiceRegistration, error),
        GetService: PrimitiveTypes.Func(ServiceKey)(PrimitiveTypes.Any, error),
    },
});
