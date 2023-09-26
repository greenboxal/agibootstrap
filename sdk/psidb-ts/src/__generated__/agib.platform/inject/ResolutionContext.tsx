import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { error } from "@psidb/psidb-sdk/types//error";
import { ServiceKey } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceKey";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { ServiceRegistration } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceRegistration";


export const ResolutionContext = makeInterface({
    name: "agib.platform/inject/ResolutionContext",
    methods: {
        AppendShutdownHook: PrimitiveTypes.Func(PrimitiveTypes.Func(Context)(error))(),
        GetRegistration: PrimitiveTypes.Func(ServiceKey, bool)(ServiceRegistration, error),
        GetService: PrimitiveTypes.Func(ServiceKey)(PrimitiveTypes.Any, error),
        Path: PrimitiveTypes.Func()(PrimitiveTypes.Array(ServiceKey)),
    },
});
