import { makeSchema, ArrayOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class Topic extends makeSchema("psidb.chat/Topic", {
    members: ArrayOf(Path),
    name: PrimitiveTypes.String,
}) {}
