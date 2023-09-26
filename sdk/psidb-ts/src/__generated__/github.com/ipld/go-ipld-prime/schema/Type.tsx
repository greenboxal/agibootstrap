import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Kind } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Kind";
import { TypeKind } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/schema/TypeKind";
import { TypeSystem } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/schema/TypeSystem";


export const Type = makeInterface({
    name: "github.com/ipld/go-ipld-prime/schema/Type",
    methods: {
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        RepresentationBehavior: PrimitiveTypes.Func()(Kind),
        TypeKind: PrimitiveTypes.Func()(TypeKind),
        TypeSystem: PrimitiveTypes.Func()(PrimitiveTypes.Pointer(TypeSystem)),
    },
});
