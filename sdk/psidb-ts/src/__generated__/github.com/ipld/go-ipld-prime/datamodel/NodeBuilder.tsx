import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { error } from "@psidb/psidb-sdk/types//error";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Link";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { ListAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/ListAssembler";
import { MapAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/MapAssembler";
import { NodePrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodePrototype";


export const NodeBuilder = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/NodeBuilder",
    methods: {
        AssignBool: PrimitiveTypes.Func(bool)(error),
        AssignBytes: PrimitiveTypes.Func(PrimitiveTypes.Array(uint8))(error),
        AssignFloat: PrimitiveTypes.Func(PrimitiveTypes.Float64)(error),
        AssignInt: PrimitiveTypes.Func(PrimitiveTypes.Integer)(error),
        AssignLink: PrimitiveTypes.Func(Link)(error),
        AssignNode: PrimitiveTypes.Func(Node)(error),
        AssignNull: PrimitiveTypes.Func()(error),
        AssignString: PrimitiveTypes.Func(PrimitiveTypes.String)(error),
        BeginList: PrimitiveTypes.Func(PrimitiveTypes.Integer)(ListAssembler, error),
        BeginMap: PrimitiveTypes.Func(PrimitiveTypes.Integer)(MapAssembler, error),
        Build: PrimitiveTypes.Func()(Node),
        Prototype: PrimitiveTypes.Func()(NodePrototype),
        Reset: PrimitiveTypes.Func()(),
    },
});
