import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { NodeAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeAssembler";
import { error } from "@psidb/psidb-sdk/types//error";
import { NodePrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodePrototype";


export const ListAssembler = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/ListAssembler",
    methods: {
        AssembleValue: PrimitiveTypes.Func()(NodeAssembler),
        Finish: PrimitiveTypes.Func()(error),
        ValuePrototype: PrimitiveTypes.Func(PrimitiveTypes.Integer)(NodePrototype),
    },
});
