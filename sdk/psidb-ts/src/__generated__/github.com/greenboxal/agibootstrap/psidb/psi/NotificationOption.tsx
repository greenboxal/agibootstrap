import { Type, defineFunction, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Notification } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Notification";


export function NotificationOption<T0 extends Type>(t0: T0) {
    return defineFunction(Notification)(ArrayOf())
}