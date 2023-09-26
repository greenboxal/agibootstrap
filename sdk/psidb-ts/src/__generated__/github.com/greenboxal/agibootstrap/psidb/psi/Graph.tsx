import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";
import { FrozenEdge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/FrozenEdge";
import { error } from "@psidb/psidb-sdk/types//error";


export const Graph = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/Graph",
    methods: {
        Add: PrimitiveTypes.Func(Node)(),
        ListNodeEdges: PrimitiveTypes.Func(Context, Path)(PrimitiveTypes.Array(PrimitiveTypes.Pointer(FrozenEdge))(error)),
        Remove: PrimitiveTypes.Func(Node)(),
        ResolveNode: PrimitiveTypes.Func(Context, Path)(Node, error),
    },
});
