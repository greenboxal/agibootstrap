import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Type } from "@psidb/psidb-sdk/types/reflect/Type";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Value } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Value";
import { error } from "@psidb/psidb-sdk/types//error";
import { Field } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Field";
import { NodePrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodePrototype";
import { TypedPrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/schema/TypedPrototype";
import { Kind } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Kind";
import { Schema } from "@psidb/psidb-sdk/types/github.com/invopop/jsonschema/Schema";
import { ListType } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/ListType";
import { MapType } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/MapType";
import { Method } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Method";
import { TypeName } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/TypeName";
import { PrimitiveKind } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/PrimitiveKind";

const _F = {} as any

export const StructType = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/typesystem/StructType",
    methods: {
        AssignableTo: PrimitiveTypes.Func(Type)(bool),
        ConvertFromAny: PrimitiveTypes.Func(Value)(Value, error),
        Field: PrimitiveTypes.Func(PrimitiveTypes.String)(Field),
        FieldByIndex: PrimitiveTypes.Func(PrimitiveTypes.Integer)(Field),
        IpldPrimitive: PrimitiveTypes.Func()(NodePrototype),
        IpldPrototype: PrimitiveTypes.Func()(TypedPrototype),
        IpldRepresentationKind: PrimitiveTypes.Func()(Kind),
        IpldType: PrimitiveTypes.Func()(Type),
        JsonSchema: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(Schema)),
        List: PrimitiveTypes.Func()(ListType),
        Map: PrimitiveTypes.Func()(MapType),
        Method: PrimitiveTypes.Func(PrimitiveTypes.String)(Method),
        MethodByIndex: PrimitiveTypes.Func(PrimitiveTypes.Integer)(Method),
        Name: PrimitiveTypes.Func()(TypeName),
        NumField: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        NumMethods: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        PrimitiveKind: PrimitiveTypes.Func()(PrimitiveKind),
        RuntimeType: PrimitiveTypes.Func()(Type),
        Struct: PrimitiveTypes.Func()(_F["StructType"]),
    },
});
_F["StructType"] = StructType;
