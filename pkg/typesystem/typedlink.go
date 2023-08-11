package typesystem

import "github.com/ipld/go-ipld-prime"

type TypedLink interface {
	ipld.Link

	LinkedObjectType() Type
}
