import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { Reader } from "@psidb/psidb-sdk/types/io/Reader";
import { Decoder } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/codec/Decoder";
import { NodeAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeAssembler";
import { error } from "@psidb/psidb-sdk/types//error";
import { NodeTypeDefinition } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/NodeTypeDefinition";
import { Writer } from "@psidb/psidb-sdk/types/io/Writer";
import { VTable } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/VTable";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Type } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Type";
import { TypeName } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/TypeName";


export const NodeType = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/NodeType",
    methods: {
        CreateInstance: PrimitiveTypes.Func()(Node),
        DecodeNode: PrimitiveTypes.Func(Reader, Decoder(NodeAssembler, Reader)(error))(Node, error),
        Definition: PrimitiveTypes.Func()(NodeTypeDefinition),
        EncodeNode: PrimitiveTypes.Func(Writer, Encoder(Node, Writer)(error, Node))(error),
        InitializeNode: PrimitiveTypes.Func(Node)(),
        Interface: PrimitiveTypes.Func(PrimitiveTypes.String)(PrimitiveTypes.Pointer(VTable)),
        Interfaces: PrimitiveTypes.Func()(PrimitiveTypes.Array(PrimitiveTypes.Pointer(VTable))),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        OnAfterNodeLoaded: PrimitiveTypes.Func(Context, Node)(error),
        OnBeforeNodeSaved: PrimitiveTypes.Func(Context, Node)(error),
        RuntimeType: PrimitiveTypes.Func()(Type),
        String: PrimitiveTypes.Func()(PrimitiveTypes.String),
        Type: PrimitiveTypes.Func()(Type),
        TypeName: PrimitiveTypes.Func()(TypeName),
    },
});
