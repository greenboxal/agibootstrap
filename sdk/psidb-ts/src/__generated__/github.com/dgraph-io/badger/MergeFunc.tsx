import { Type, defineFunction, SequenceOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export function MergeFunc<T0 extends Type>(t0: T0) {
    return defineFunction(SequenceOf(PrimitiveTypes.Array(uint8), PrimitiveTypes.Array(uint8)))(PrimitiveTypes.Array(uint8))
}