import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { EdgeKey } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeKey";
import { EdgeKind } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeKind";


export const EdgeReference = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/EdgeReference",
    methods: {
        GetIndex: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        GetKey: PrimitiveTypes.Func()(EdgeKey),
        GetKind: PrimitiveTypes.Func()(EdgeKind),
        GetName: PrimitiveTypes.Func()(PrimitiveTypes.String),
    },
});
