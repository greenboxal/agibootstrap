import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export const error = makeInterface({
    name: "error",
    methods: {
        Error: PrimitiveTypes.Func()(PrimitiveTypes.String),
    },
});
