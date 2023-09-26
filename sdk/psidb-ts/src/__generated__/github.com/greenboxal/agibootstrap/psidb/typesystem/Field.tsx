import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { StructType } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/StructType";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { Value } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Value";
import { StructField } from "@psidb/psidb-sdk/types/reflect/StructField";
import { Type } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/typesystem/Type";


export const Field = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/typesystem/Field",
    methods: {
        DeclaringType: PrimitiveTypes.Func()(StructType),
        IsNullable: PrimitiveTypes.Func()(bool),
        IsOptional: PrimitiveTypes.Func()(bool),
        IsVirtual: PrimitiveTypes.Func()(bool),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        Resolve: PrimitiveTypes.Func(Value)(Value),
        RuntimeField: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(StructField)),
        Type: PrimitiveTypes.Func()(Type),
    },
});
