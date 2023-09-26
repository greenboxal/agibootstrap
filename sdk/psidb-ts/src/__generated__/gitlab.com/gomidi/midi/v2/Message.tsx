import { Type, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export function Message<T0 extends Type>(t0: T0) {
    return ArrayOf(uint8)
}