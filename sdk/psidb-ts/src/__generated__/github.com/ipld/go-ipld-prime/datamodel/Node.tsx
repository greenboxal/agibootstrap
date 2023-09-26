import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { error } from "@psidb/psidb-sdk/types//error";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Link";
import { Kind } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Kind";
import { ListIterator } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/ListIterator";
import { PathSegment } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/PathSegment";
import { MapIterator } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/MapIterator";
import { NodePrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodePrototype";

const _F = {} as any

export const Node = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/Node",
    methods: {
        AsBool: PrimitiveTypes.Func()(bool, error),
        AsBytes: PrimitiveTypes.Func()(PrimitiveTypes.Array(uint8)(error)),
        AsFloat: PrimitiveTypes.Func()(PrimitiveTypes.Float64, error),
        AsInt: PrimitiveTypes.Func()(PrimitiveTypes.Integer, error),
        AsLink: PrimitiveTypes.Func()(Link, error),
        AsString: PrimitiveTypes.Func()(PrimitiveTypes.String, error),
        IsAbsent: PrimitiveTypes.Func()(bool),
        IsNull: PrimitiveTypes.Func()(bool),
        Kind: PrimitiveTypes.Func()(Kind),
        Length: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        ListIterator: PrimitiveTypes.Func()(ListIterator),
        LookupByIndex: PrimitiveTypes.Func(PrimitiveTypes.Integer)(_F["Node"], error),
        LookupByNode: PrimitiveTypes.Func(_F["Node"])(_F["Node"], error),
        LookupBySegment: PrimitiveTypes.Func(PathSegment)(_F["Node"], error),
        LookupByString: PrimitiveTypes.Func(PrimitiveTypes.String)(_F["Node"], error),
        MapIterator: PrimitiveTypes.Func()(MapIterator),
        Prototype: PrimitiveTypes.Func()(NodePrototype),
    },
});
_F["Node"] = Node;
