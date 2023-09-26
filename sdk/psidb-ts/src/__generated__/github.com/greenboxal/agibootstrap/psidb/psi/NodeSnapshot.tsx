import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Link";
import { FrozenEdge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/FrozenEdge";
import { FrozenNode } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/FrozenNode";
import { EdgeReference } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeReference";
import { Edge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Edge";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { error } from "@psidb/psidb-sdk/types//error";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export const NodeSnapshot = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/NodeSnapshot",
    methods: {
        CommitLink: PrimitiveTypes.Func()(Link),
        CommitVersion: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        FrozenEdges: PrimitiveTypes.Func()(PrimitiveTypes.Array(PrimitiveTypes.Pointer(FrozenEdge))),
        FrozenNode: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(FrozenNode)),
        ID: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        LastFenceID: PrimitiveTypes.Func()(PrimitiveTypes.UnsignedInteger),
        Lookup: PrimitiveTypes.Func(EdgeReference)(Edge),
        Node: PrimitiveTypes.Func()(Node),
        OnAfterInitialize: PrimitiveTypes.Func(Node)(),
        OnAttributeChanged: PrimitiveTypes.Func(PrimitiveTypes.String, PrimitiveTypes.Any)(),
        OnAttributeRemoved: PrimitiveTypes.Func(PrimitiveTypes.String, PrimitiveTypes.Any)(),
        OnBeforeInitialize: PrimitiveTypes.Func(Node)(),
        OnChildAdded: PrimitiveTypes.Func(Node)(),
        OnChildRemoved: PrimitiveTypes.Func(Node)(),
        OnEdgeAdded: PrimitiveTypes.Func(Edge)(),
        OnEdgeRemoved: PrimitiveTypes.Func(Edge)(),
        OnInvalidated: PrimitiveTypes.Func()(),
        OnParentChange: PrimitiveTypes.Func(Node)(),
        OnUpdated: PrimitiveTypes.Func(Context)(error),
        Path: PrimitiveTypes.Func()(Path),
        Resolve: PrimitiveTypes.Func(Context, Path)(Node, error),
    },
});
