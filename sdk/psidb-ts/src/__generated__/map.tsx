import { Type, MapOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";


export function map<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return MapOf(PrimitiveTypes.String, t1)
}

















