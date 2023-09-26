import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { map } from "@psidb/psidb-sdk/types//map";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";
import { NodeIterator } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/NodeIterator";
import { ObservableList } from "@psidb/psidb-sdk/types/agib.platform/stdlib/obsfx/collectionsfx/ObservableList";
import { EdgeIterator } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeIterator";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { EdgeReference } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeReference";
import { Edge } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Edge";
import { NodeBase } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/NodeBase";
import { NodeType } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/NodeType";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { PathElement } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/PathElement";
import { error } from "@psidb/psidb-sdk/types//error";


export const UniqueNode = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/UniqueNode",
    methods: {
        AddChildNode: PrimitiveTypes.Func(Node)(),
        Attributes: PrimitiveTypes.Func()(map(PrimitiveTypes.String, PrimitiveTypes.Any)),
        CanonicalPath: PrimitiveTypes.Func()(Path),
        Children: PrimitiveTypes.Func()(PrimitiveTypes.Array(Node)),
        ChildrenIterator: PrimitiveTypes.Func()(NodeIterator),
        ChildrenList: PrimitiveTypes.Func()(ObservableList(Node)),
        Comments: PrimitiveTypes.Func()(PrimitiveTypes.Array(PrimitiveTypes.String)),
        Edges: PrimitiveTypes.Func()(EdgeIterator),
        GetAttribute: PrimitiveTypes.Func(PrimitiveTypes.String)(PrimitiveTypes.Any, bool),
        GetEdge: PrimitiveTypes.Func(EdgeReference)(Edge),
        ID: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        InsertChildAfter: PrimitiveTypes.Func(Node, Node)(),
        InsertChildBefore: PrimitiveTypes.Func(Node, Node)(),
        InsertChildrenAt: PrimitiveTypes.Func(PrimitiveTypes.Integer, Node)(),
        Invalidate: PrimitiveTypes.Func()(),
        IsContainer: PrimitiveTypes.Func()(bool),
        IsLeaf: PrimitiveTypes.Func()(bool),
        IsValid: PrimitiveTypes.Func()(bool),
        NextSibling: PrimitiveTypes.Func()(Node),
        Parent: PrimitiveTypes.Func()(Node),
        PreviousSibling: PrimitiveTypes.Func()(Node),
        PsiNode: PrimitiveTypes.Func()(Node),
        PsiNodeBase: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(NodeBase)),
        PsiNodeType: PrimitiveTypes.Func()(NodeType),
        PsiNodeVersion: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        RemoveAttribute: PrimitiveTypes.Func(PrimitiveTypes.String)(PrimitiveTypes.Any, bool),
        RemoveChildNode: PrimitiveTypes.Func(Node)(),
        ReplaceChildNode: PrimitiveTypes.Func(Node, Node)(),
        ResolveChild: PrimitiveTypes.Func(Context, PathElement)(Node),
        SelfIdentity: PrimitiveTypes.Func()(Path),
        SetAttribute: PrimitiveTypes.Func(PrimitiveTypes.String, PrimitiveTypes.Any)(),
        SetEdge: PrimitiveTypes.Func(EdgeReference, Node)(),
        SetParent: PrimitiveTypes.Func(Node)(),
        String: PrimitiveTypes.Func()(PrimitiveTypes.String),
        UUID: PrimitiveTypes.Func()(PrimitiveTypes.String),
        UnsetEdge: PrimitiveTypes.Func(EdgeReference)(),
        Update: PrimitiveTypes.Func(Context)(error),
        UpsertEdge: PrimitiveTypes.Func(Edge)(),
    },
});
