import { Type, defineFunction } from "@psidb/psidb-sdk/client/schema";
import { TaskProgress } from "@psidb/psidb-sdk/types/agib.platform/tasks/TaskProgress";
import { error } from "@psidb/psidb-sdk/types//error";


export function TaskFunc<T0 extends Type>(t0: T0) {
    return defineFunction(TaskProgress)(error)
}