import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Iterator } from "@psidb/psidb-sdk/types/agib.platform/stdlib/iterators/Iterator";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";
import { Thought } from "@psidb/psidb-sdk/types/agib.platform/db/thoughtdb/Thought";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Pointer } from "@psidb/psidb-sdk/types/agib.platform/db/thoughtdb/Pointer";
import { error } from "@psidb/psidb-sdk/types//error";
import { WalkFunc } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/WalkFunc";
import { Cursor } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Cursor";

const _F = {} as any

export const Cursor = makeInterface({
    name: "agib.platform/db/thoughtdb/Cursor",
    methods: {
        Clone: PrimitiveTypes.Func()(_F["Cursor"]),
        Depth: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Enqueue: PrimitiveTypes.Func(Iterator(Node))(),
        EnqueueParents: PrimitiveTypes.Func()(),
        InsertAfter: PrimitiveTypes.Func(Node)(),
        InsertBefore: PrimitiveTypes.Func(Node)(),
        IterateParents: PrimitiveTypes.Func()(Iterator(PrimitiveTypes.Pointer(Thought))),
        Next: PrimitiveTypes.Func()(bool),
        Pointer: PrimitiveTypes.Func()(Pointer),
        Pop: PrimitiveTypes.Func()(bool),
        Push: PrimitiveTypes.Func(Iterator(Node))(),
        PushChildren: PrimitiveTypes.Func()(),
        PushEdges: PrimitiveTypes.Func()(),
        PushParents: PrimitiveTypes.Func()(),
        PushPointer: PrimitiveTypes.Func(Pointer)(error),
        PushThought: PrimitiveTypes.Func(PrimitiveTypes.Pointer(Thought))(),
        Replace: PrimitiveTypes.Func(Node)(),
        SetCurrent: PrimitiveTypes.Func(Node)(),
        SetNext: PrimitiveTypes.Func(Node)(),
        SkipChildren: PrimitiveTypes.Func()(),
        SkipEdges: PrimitiveTypes.Func()(),
        Thought: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(Thought)),
        Value: PrimitiveTypes.Func()(Node),
        Walk: PrimitiveTypes.Func(Node, WalkFunc(Cursor, bool)(error))(error),
        WalkChildren: PrimitiveTypes.Func()(),
        WalkEdges: PrimitiveTypes.Func()(),
    },
});
_F["Cursor"] = Cursor;
