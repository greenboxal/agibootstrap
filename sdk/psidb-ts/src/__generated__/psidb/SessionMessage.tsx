import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { SessionMessageHeader } from "@psidb/psidb-sdk/types/psidb/SessionMessageHeader";


export const SessionMessage = makeInterface({
    name: "psidb/SessionMessage",
    methods: {
        SessionMessageHeader: PrimitiveTypes.Func()(SessionMessageHeader),
        SessionMessageMarker: PrimitiveTypes.Func()(),
        SetSessionMessageHeader: PrimitiveTypes.Func(SessionMessageHeader)(),
    },
});
