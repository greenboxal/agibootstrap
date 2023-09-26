import { makeSchema, MapOf, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { NodeReifier } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/NodeReifier";
import { LinkContext } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/linking/LinkContext";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { error } from "@psidb/psidb-sdk/types//error";

const _F = {} as any

export class LinkSystem extends makeSchema("github.com/ipld/go-ipld-prime/linking/LinkSystem", {
    KnownReifiers: MapOf(PrimitiveTypes.String, NodeReifier(LinkContext, Node, PrimitiveTypes.Pointer(_F["LinkSystem"]))(Node, error)),
    TrustedStorage: PrimitiveTypes.Boolean,
}) {}
_F["LinkSystem"] = LinkSystem;
