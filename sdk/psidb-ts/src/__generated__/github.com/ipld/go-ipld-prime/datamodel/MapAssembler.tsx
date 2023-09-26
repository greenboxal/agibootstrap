import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { NodeAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeAssembler";
import { error } from "@psidb/psidb-sdk/types//error";
import { NodePrototype } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodePrototype";


export const MapAssembler = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/MapAssembler",
    methods: {
        AssembleEntry: PrimitiveTypes.Func(PrimitiveTypes.String)(NodeAssembler, error),
        AssembleKey: PrimitiveTypes.Func()(NodeAssembler),
        AssembleValue: PrimitiveTypes.Func()(NodeAssembler),
        Finish: PrimitiveTypes.Func()(error),
        KeyPrototype: PrimitiveTypes.Func()(NodePrototype),
        ValuePrototype: PrimitiveTypes.Func(PrimitiveTypes.String)(NodePrototype),
    },
});
