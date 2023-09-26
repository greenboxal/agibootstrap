import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { error } from "@psidb/psidb-sdk/types//error";
import { ServiceKey } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceKey";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { ServiceRegistration } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceRegistration";
import { ServiceDefinition } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceDefinition";


export const ServiceProvider = makeInterface({
    name: "agib.platform/inject/ServiceProvider",
    methods: {
        AppendShutdownHook: PrimitiveTypes.Func(PrimitiveTypes.Func(Context)(error))(),
        Close: PrimitiveTypes.Func(Context)(error),
        GetRegistration: PrimitiveTypes.Func(ServiceKey, bool)(ServiceRegistration, error),
        GetService: PrimitiveTypes.Func(ServiceKey)(PrimitiveTypes.Any, error),
        RegisterService: PrimitiveTypes.Func(ServiceDefinition)(),
    },
});
