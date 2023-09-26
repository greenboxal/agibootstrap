import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";
import { error } from "@psidb/psidb-sdk/types//error";


export const GraphOperations = makeInterface({
    name: "psidb/GraphOperations",
    methods: {
        Add: PrimitiveTypes.Func(Node)(),
        Remove: PrimitiveTypes.Func(Node)(),
        Resolve: PrimitiveTypes.Func(Context, Path)(Node, error),
    },
});
