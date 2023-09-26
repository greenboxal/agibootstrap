import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Iterator } from "@psidb/psidb-sdk/types/agib.platform/stdlib/iterators/Iterator";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { WalkFunc } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/WalkFunc";
import { error } from "@psidb/psidb-sdk/types//error";

const _F = {} as any

export const Cursor = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/Cursor",
    methods: {
        Depth: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Enqueue: PrimitiveTypes.Func(Iterator(Node))(),
        InsertAfter: PrimitiveTypes.Func(Node)(),
        InsertBefore: PrimitiveTypes.Func(Node)(),
        Next: PrimitiveTypes.Func()(bool),
        Pop: PrimitiveTypes.Func()(bool),
        Push: PrimitiveTypes.Func(Iterator(Node))(),
        PushChildren: PrimitiveTypes.Func()(),
        PushEdges: PrimitiveTypes.Func()(),
        Replace: PrimitiveTypes.Func(Node)(),
        SetCurrent: PrimitiveTypes.Func(Node)(),
        SetNext: PrimitiveTypes.Func(Node)(),
        SkipChildren: PrimitiveTypes.Func()(),
        SkipEdges: PrimitiveTypes.Func()(),
        Value: PrimitiveTypes.Func()(Node),
        Walk: PrimitiveTypes.Func(Node, WalkFunc(_F["Cursor"], bool)(error))(error),
        WalkChildren: PrimitiveTypes.Func()(),
        WalkEdges: PrimitiveTypes.Func()(),
    },
});
_F["Cursor"] = Cursor;
