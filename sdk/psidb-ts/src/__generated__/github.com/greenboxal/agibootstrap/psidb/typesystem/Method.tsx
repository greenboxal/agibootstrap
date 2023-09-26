import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Value } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Value";
import { error } from "@psidb/psidb-sdk/types//error";
import { Type } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Type";
import { bool } from "@psidb/psidb-sdk/types//bool";


export const Method = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/typesystem/Method",
    methods: {
        Call: PrimitiveTypes.Func(Value, PrimitiveTypes.Array(Value))(PrimitiveTypes.Array(Value)(error)),
        CallSlice: PrimitiveTypes.Func(Value, PrimitiveTypes.Array(Value))(PrimitiveTypes.Array(Value)(error)),
        DeclaringType: PrimitiveTypes.Func()(Type),
        In: PrimitiveTypes.Func(PrimitiveTypes.Integer)(Type),
        IsVariadic: PrimitiveTypes.Func()(bool),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        NumIn: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        NumOut: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        Out: PrimitiveTypes.Func(PrimitiveTypes.Integer)(Type),
    },
});
