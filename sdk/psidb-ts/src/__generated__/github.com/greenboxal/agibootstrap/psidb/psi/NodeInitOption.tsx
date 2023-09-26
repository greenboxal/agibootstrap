import { Type, defineFunction, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { NodeBase } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/NodeBase";


export function NodeInitOption<T0 extends Type>(t0: T0) {
    return defineFunction(NodeBase)(ArrayOf())
}