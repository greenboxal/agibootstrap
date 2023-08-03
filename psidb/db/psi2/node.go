package psi2

type NodeLike interface {
	PsiNodeBase() *NodeBase
	PsiNodeSnapshot() NodeSnapshot
}

type Node interface {
	NodeLike

	GetParent() Node
	SetParent(parent Node)
}

type UniqueNode interface {
	Node

	UUID() string
}
