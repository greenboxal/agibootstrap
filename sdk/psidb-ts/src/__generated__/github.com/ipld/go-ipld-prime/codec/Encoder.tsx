import { Type, defineFunction, SequenceOf } from "@psidb/psidb-sdk/client/schema";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { Writer } from "@psidb/psidb-sdk/types/io/Writer";
import { error } from "@psidb/psidb-sdk/types//error";


export function Encoder<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return defineFunction(SequenceOf(Node, Writer))(error)
}