import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";
import { Cid } from "@psidb/psidb-sdk/types/github.com/ipfs/go-cid/Cid";
import { error } from "@psidb/psidb-sdk/types//error";

const _F = {} as any

export const Builder = makeInterface({
    name: "github.com/ipfs/go-cid/Builder",
    methods: {
        GetCodec: PrimitiveTypes.Func()(PrimitiveTypes.UnsignedInteger),
        Sum: PrimitiveTypes.Func(PrimitiveTypes.Array(uint8))(Cid, error),
        WithCodec: PrimitiveTypes.Func(PrimitiveTypes.UnsignedInteger)(_F["Builder"]),
    },
});
_F["Builder"] = Builder;
