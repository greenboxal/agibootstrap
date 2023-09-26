import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export const MessageV1 = makeInterface({
    name: "google.golang.org/protobuf/runtime/protoiface/MessageV1",
    methods: {
        ProtoMessage: PrimitiveTypes.Func()(),
        Reset: PrimitiveTypes.Func()(),
        String: PrimitiveTypes.Func()(PrimitiveTypes.String),
    },
});
