import { Type, defineFunction, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { ExecuteOptions } from "@psidb/psidb-sdk/types/psidb.agents/ExecuteOptions";


export function ExecuteOption<T0 extends Type>(t0: T0) {
    return defineFunction(ExecuteOptions)(ArrayOf())
}