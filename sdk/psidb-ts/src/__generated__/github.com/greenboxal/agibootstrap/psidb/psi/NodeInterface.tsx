import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { NodeActionDefinition } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/NodeActionDefinition";
import { VTableDefinition } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/VTableDefinition";
import { error } from "@psidb/psidb-sdk/types//error";


export const NodeInterface = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/psi/NodeInterface",
    methods: {
        Actions: PrimitiveTypes.Func()(PrimitiveTypes.Array(NodeActionDefinition)),
        Description: PrimitiveTypes.Func()(PrimitiveTypes.String),
        Name: PrimitiveTypes.Func()(PrimitiveTypes.String),
        ValidateImplementation: PrimitiveTypes.Func(VTableDefinition)(error),
    },
});
