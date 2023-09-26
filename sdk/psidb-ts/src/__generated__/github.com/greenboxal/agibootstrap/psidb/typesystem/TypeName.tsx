import { makeSchema, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";

const _F = {} as any

export class TypeName extends makeSchema("github.com/greenboxal/agibootstrap/psidb/typesystem/TypeName", {
    Class: PrimitiveTypes.Float64,
    InParameters: ArrayOf(_F["TypeName"]),
    Name: PrimitiveTypes.String,
    OutParameters: ArrayOf(_F["TypeName"]),
    Package: PrimitiveTypes.String,
}) {}
_F["TypeName"] = TypeName;
