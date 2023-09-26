import { Type, defineFunction, SequenceOf } from "@psidb/psidb-sdk/client/schema";
import { Cursor } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Cursor";
import { bool } from "@psidb/psidb-sdk/types//bool";
import { error } from "@psidb/psidb-sdk/types//error";


export function WalkFunc<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return defineFunction(SequenceOf(Cursor, bool))(error)
}