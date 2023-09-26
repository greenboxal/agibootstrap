import { Type, defineFunction, SequenceOf } from "@psidb/psidb-sdk/client/schema";
import { LinkContext } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/LinkContext";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { LinkSystem } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/LinkSystem";
import { error } from "@psidb/psidb-sdk/types//error";


export function NodeReifier<T0 extends Type, T1 extends Type, T2 extends Type>(t0: T0, t1: T1, t2: T2) {
    return defineFunction(SequenceOf(LinkContext, Node, LinkSystem))(SequenceOf(Node, error))
}