import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { EdgeKind } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeKind";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Graph } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Graph";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { EdgeKey } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeKey";
import { error } from "@psidb/psidb-sdk/types//error";


export const EdgeType = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/EdgeType",
    methods: {
        IsCachedBasedOnFrom: PrimitiveTypes.Func()(bool),
        IsIndexed: PrimitiveTypes.Func()(bool),
        IsNamed: PrimitiveTypes.Func()(bool),
        IsVirtual: PrimitiveTypes.Func()(bool),
        Kind: PrimitiveTypes.Func()(EdgeKind),
        Resolve: PrimitiveTypes.Func(Context, Graph, Node, EdgeKey)(Node, error),
    },
});
