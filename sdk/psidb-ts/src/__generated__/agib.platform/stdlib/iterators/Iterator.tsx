import { Type, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { BasicSearchHit } from "@psidb/psidb-sdk/types/psidb.indexing/BasicSearchHit";


export function Iterator<T0 extends Type>(t0: T0) {
    return makeInterface({
        name: "agib.platform/stdlib/iterators/Iterator(psidb.indexing/BasicSearchHit)",
        methods: {
            Next: PrimitiveTypes.Func()(bool),
            Value: PrimitiveTypes.Func()(BasicSearchHit),
        },
    })
}













