import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Link } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Link";


export const LinkPrototype = makeInterface({
    name: "github.com/ipld/go-ipld-prime/datamodel/LinkPrototype",
    methods: {
        BuildLink: PrimitiveTypes.Func(PrimitiveTypes.Array(uint8))(Link),
    },
});
