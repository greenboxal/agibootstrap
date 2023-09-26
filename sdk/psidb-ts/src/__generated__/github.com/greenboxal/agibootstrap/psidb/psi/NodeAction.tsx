import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { error } from "@psidb/psidb-sdk/types//error";
import { Type } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Type";


export const NodeAction = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/NodeAction",
    methods: {
        Invoke: PrimitiveTypes.Func(Context, Node, PrimitiveTypes.Any)(PrimitiveTypes.Any, error),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        RequestType: PrimitiveTypes.Func()(Type),
        ResponseType: PrimitiveTypes.Func()(Type),
    },
});
