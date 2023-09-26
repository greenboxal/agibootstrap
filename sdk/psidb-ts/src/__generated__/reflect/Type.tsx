import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { ChanDir } from "@psidb/psidb-sdk/types/reflect/ChanDir";
import { StructField } from "@psidb/psidb-sdk/types/reflect/StructField";
import { Kind } from "@psidb/psidb-sdk/types/reflect/Kind";
import { Method } from "@psidb/psidb-sdk/types/reflect/Method";
import { uintptr } from "@psidb/psidb-sdk/types//uintptr";

const _F = {} as any

export const Type = makeInterface({
    name: "reflect/Type",
    methods: {
        Align: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        AssignableTo: PrimitiveTypes.Func(_F["Type"])(bool),
        Bits: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        ChanDir: PrimitiveTypes.Func()(ChanDir),
        Comparable: PrimitiveTypes.Func()(bool),
        ConvertibleTo: PrimitiveTypes.Func(_F["Type"])(bool),
        Elem: PrimitiveTypes.Func()(_F["Type"]),
        Field: PrimitiveTypes.Func(PrimitiveTypes.Integer)(StructField),
        FieldAlign: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        FieldByIndex: PrimitiveTypes.Func(PrimitiveTypes.Array(PrimitiveTypes.Integer))(StructField),
        FieldByName: PrimitiveTypes.Func(PrimitiveTypes.String)(StructField, bool),
        FieldByNameFunc: PrimitiveTypes.Func(PrimitiveTypes.Func(PrimitiveTypes.String)(bool))(StructField, bool),
        Implements: PrimitiveTypes.Func(_F["Type"])(bool),
        In: PrimitiveTypes.Func(PrimitiveTypes.Integer)(_F["Type"]),
        IsVariadic: PrimitiveTypes.Func()(bool),
        Key: PrimitiveTypes.Func()(_F["Type"]),
        Kind: PrimitiveTypes.Func()(Kind),
        Len: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Method: PrimitiveTypes.Func(PrimitiveTypes.Integer)(Method),
        MethodByName: PrimitiveTypes.Func(PrimitiveTypes.String)(Method, bool),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        NumField: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        NumIn: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        NumMethod: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        NumOut: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Out: PrimitiveTypes.Func(PrimitiveTypes.Integer)(_F["Type"]),
        PkgPath: PrimitiveTypes.Func()(PrimitiveTypes.String),
        Size: PrimitiveTypes.Func()(uintptr),
        String: PrimitiveTypes.Func()(PrimitiveTypes.String),
    },
});
_F["Type"] = Type;
