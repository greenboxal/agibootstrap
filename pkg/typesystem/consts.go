package typesystem

type PrimitiveKind int

const (
	PrimitiveKindInvalid PrimitiveKind = iota
	PrimitiveKindBoolean
	PrimitiveKindBytes
	PrimitiveKindString
	PrimitiveKindInt
	PrimitiveKindUnsignedInt
	PrimitiveKindFloat
	PrimitiveKindList
	PrimitiveKindMap
	PrimitiveKindStruct
	PrimitiveKindInterface
	PrimitiveKindLink
)
