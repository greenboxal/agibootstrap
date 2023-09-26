import { Type, defineFunction, SequenceOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { PromptBuilder } from "@psidb/psidb-sdk/types/psidb.agents/PromptBuilder";
import { Iterator } from "@psidb/psidb-sdk/types/agib.platform/stdlib/iterators/Iterator";
import { Message } from "@psidb/psidb-sdk/types/psidb.chat/Message";
import { error } from "@psidb/psidb-sdk/types//error";


export function PromptMessageSource<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return defineFunction(SequenceOf(Context, PromptBuilder))(SequenceOf(Iterator(PrimitiveTypes.Pointer(Message)), error))
}