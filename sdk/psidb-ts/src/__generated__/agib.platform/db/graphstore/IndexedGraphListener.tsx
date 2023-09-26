import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Node";


export const IndexedGraphListener = makeInterface({
    name: "agib.platform/db/graphstore/IndexedGraphListener",
    methods: {
        OnNodeUpdated: PrimitiveTypes.Func(Node)(),
    },
});
