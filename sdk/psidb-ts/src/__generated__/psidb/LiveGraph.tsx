import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { error } from "@psidb/psidb-sdk/types//error";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";
import { FrozenEdge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/FrozenEdge";
import { ServiceLocator } from "@psidb/psidb-sdk/types/agib.platform/inject/ServiceLocator";


export const LiveGraph = makeInterface({
    name: "psidb/LiveGraph",
    methods: {
        Add: PrimitiveTypes.Func(Node)(),
        Delete: PrimitiveTypes.Func(Context, Node)(error),
        ListNodeEdges: PrimitiveTypes.Func(Context, Path)(PrimitiveTypes.Array(PrimitiveTypes.Pointer(FrozenEdge))(error)),
        Remove: PrimitiveTypes.Func(Node)(),
        Resolve: PrimitiveTypes.Func(Context, Path)(Node, error),
        ResolveNode: PrimitiveTypes.Func(Context, Path)(Node, error),
        Root: PrimitiveTypes.Func()(Path),
        Services: PrimitiveTypes.Func()(ServiceLocator),
    },
});
