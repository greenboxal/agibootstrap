import { Type, defineFunction, SequenceOf } from "@psidb/psidb-sdk/client/schema";
import { NodeAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeAssembler";
import { Reader } from "@psidb/psidb-sdk/types/io/Reader";
import { error } from "@psidb/psidb-sdk/types//error";


export function Decoder<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return defineFunction(SequenceOf(NodeAssembler, Reader))(error)
}