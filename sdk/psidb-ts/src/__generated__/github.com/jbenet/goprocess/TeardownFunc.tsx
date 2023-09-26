import { defineFunction, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";


export const TeardownFunc = defineFunction(ArrayOf())(error);
