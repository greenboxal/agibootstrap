import { makeSchema } from "@psidb/psidb-sdk/client/schema";
import { Context } from "@psidb/psidb-sdk/types/context/Context";
import { Node } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/Node";
import { NodeAssembler } from "@psidb/psidb-sdk/types/github.com/ipld/go-ipld-prime/datamodel/NodeAssembler";


export class LinkContext extends makeSchema("github.com/ipld/go-ipld-prime/linking/LinkContext", {
    Ctx: Context,
    LinkNode: Node,
    LinkNodeAssembler: NodeAssembler,
    LinkPath: makeSchema("", {
    }),
    ParentNode: Node,
}) {}
