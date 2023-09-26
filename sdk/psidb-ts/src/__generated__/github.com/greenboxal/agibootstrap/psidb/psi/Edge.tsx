import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { EdgeID } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeID";
import { EdgeReference } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeReference";
import { EdgeKind } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeKind";
import { EdgeBase } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/EdgeBase";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { error } from "@psidb/psidb-sdk/types//error";

const _F = {} as any

export const Edge = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/Edge",
    methods: {
        From: PrimitiveTypes.Func()(Node),
        ID: PrimitiveTypes.Func()(EdgeID),
        Key: PrimitiveTypes.Func()(EdgeReference),
        Kind: PrimitiveTypes.Func()(EdgeKind),
        PsiEdgeBase: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(EdgeBase)),
        ReplaceTo: PrimitiveTypes.Func(Node)(_F["Edge"]),
        ResolveTo: PrimitiveTypes.Func(Context)(Node, error),
        To: PrimitiveTypes.Func()(Node),
    },
});
_F["Edge"] = Edge;
