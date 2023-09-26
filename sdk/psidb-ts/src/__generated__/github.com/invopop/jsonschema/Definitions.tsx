import { Type, MapOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Schema } from "@psidb/psidb-sdk/types/github.com/invopop/jsonschema/Schema";


export function Definitions<T0 extends Type, T1 extends Type>(t0: T0, t1: T1) {
    return MapOf(PrimitiveTypes.String, Schema)
}