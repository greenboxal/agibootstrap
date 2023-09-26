import { makeSchema, MapOf, PrimitiveTypes, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { Path } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/psi/Path";


export class Connection extends makeSchema("psidb.kb/Connection", {
    f_score: MapOf(PrimitiveTypes.String, PrimitiveTypes.Float64),
    from: PrimitiveTypes.String,
    frontier: ArrayOf(Path),
    g_score: MapOf(PrimitiveTypes.String, PrimitiveTypes.Float64),
    links: MapOf(PrimitiveTypes.String, PrimitiveTypes.Array(Path)),
    name: PrimitiveTypes.String,
    root: PrimitiveTypes.String,
    to: PrimitiveTypes.String,
}) {}
