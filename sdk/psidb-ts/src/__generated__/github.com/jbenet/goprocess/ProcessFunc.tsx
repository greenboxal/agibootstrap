import { Type, defineFunction, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Process } from "@psidb/psidb-sdk/types/github.com/jbenet/goprocess/Process";


export function ProcessFunc<T0 extends Type>(t0: T0) {
    return defineFunction(Process)(ArrayOf())
}